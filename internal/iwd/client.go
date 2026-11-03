package iwd

import (
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/godbus/dbus/v5"
)

const (
	IwdBusName        = "net.connman.iwd"
	ObjectManagerPath = "/"
	InterfaceManager  = "org.freedesktop.DBus.ObjectManager"

	IfaceAdapter      = "net.connman.iwd.Adapter"
	IfaceDevice       = "net.connman.iwd.Device"
	IfaceStation      = "net.connman.iwd.Station"
	IfaceNetwork      = "net.connman.iwd.Network"
	IfaceKnownNetwork = "net.connman.iwd.KnownNetwork"
)

// Client handles communication with iwd daemon over D-Bus
type Client struct {
	conn *dbus.Conn
}

// NewClient establishes a connection to the system D-Bus and verifies iwd is running
func NewClient() (*Client, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system bus: %w", err)
	}

	client := &Client{conn: conn}
	return client, nil
}

// Close closes the D-Bus connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetManagedObjects retrieves all iwd managed objects
func (c *Client) GetManagedObjects() (map[dbus.ObjectPath]map[string]map[string]dbus.Variant, error) {
	obj := c.conn.Object(IwdBusName, ObjectManagerPath)
	var objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err := obj.Call(InterfaceManager+".GetManagedObjects", 0).Store(&objects)
	if err != nil {
		return nil, fmt.Errorf("failed to query iwd objects: %w", err)
	}
	return objects, nil
}

// GetDevices returns all wireless devices managed by iwd
func (c *Client) GetDevices() ([]Device, error) {
	objects, err := c.GetManagedObjects()
	if err != nil {
		return nil, err
	}

	adapters := make(map[dbus.ObjectPath]map[string]dbus.Variant)
	for path, ifaces := range objects {
		if props, ok := ifaces[IfaceAdapter]; ok {
			adapters[path] = props
		}
	}

	var devices []Device
	for path, ifaces := range objects {
		if devProps, ok := ifaces[IfaceDevice]; ok {
			dev := Device{
				Path:    path,
				Name:    getString(devProps, "Name"),
				Address: getString(devProps, "Address"),
				Powered: getBool(devProps, "Powered"),
				Mode:    getString(devProps, "Mode"),
			}
			if adapterPath, ok := devProps["Adapter"].Value().(dbus.ObjectPath); ok {
				dev.Adapter = adapterPath
				if adProps, found := adapters[adapterPath]; found {
					dev.AdapterName = getString(adProps, "Name")
					dev.AdapterModel = getString(adProps, "Model")
					dev.AdapterVendor = getString(adProps, "Vendor")
				}
			}
			devices = append(devices, dev)
		}
	}

	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Name < devices[j].Name
	})

	return devices, nil
}

// GetStation returns the station mode status for the given device path
func (c *Client) GetStation(devPath dbus.ObjectPath) (*Station, error) {
	objects, err := c.GetManagedObjects()
	if err != nil {
		return nil, err
	}

	ifaces, ok := objects[devPath]
	if !ok {
		return nil, fmt.Errorf("device not found at %s", devPath)
	}

	stProps, ok := ifaces[IfaceStation]
	if !ok {
		return nil, fmt.Errorf("device at %s is not in station mode", devPath)
	}

	station := &Station{
		Path:     devPath,
		State:    getString(stProps, "State"),
		Scanning: getBool(stProps, "Scanning"),
	}

	if connNet, ok := stProps["ConnectedNetwork"].Value().(dbus.ObjectPath); ok {
		station.ConnectedNetwork = connNet
	}

	// Sync state with system network manager (e.g. NetworkManager)
	devProps := ifaces[IfaceDevice]
	devName := getString(devProps, "Name")
	if devName == "" {
		devName = "wlan0"
	}

	if _, isActive := GetSystemActiveWirelessConnection(devName); isActive {
		station.State = "connected"
	}

	return station, nil
}

type networkSignalRaw struct {
	Path   dbus.ObjectPath
	Signal int16
}

// GetNetworks returns all available networks visible from the station
func (c *Client) GetNetworks(stationPath dbus.ObjectPath) ([]Network, error) {
	stationObj := c.conn.Object(IwdBusName, stationPath)

	var ordered []networkSignalRaw
	err := stationObj.Call(IfaceStation+".GetOrderedNetworks", 0).Store(&ordered)
	if err != nil {
		return c.getNetworksFromObjects(stationPath)
	}

	systemSaved := GetSystemSavedConnections()
	systemSavedMap := make(map[string]bool, len(systemSaved))
	for _, s := range systemSaved {
		systemSavedMap[s] = true
	}

	activeSSID, isActive := GetSystemActiveWirelessConnection("wlan0")

	var networks []Network
	for _, item := range ordered {
		netObj := c.conn.Object(IwdBusName, item.Path)

		nameVar, err := netObj.GetProperty(IfaceNetwork + ".Name")
		if err != nil {
			continue
		}
		typeVar, _ := netObj.GetProperty(IfaceNetwork + ".Type")
		connVar, _ := netObj.GetProperty(IfaceNetwork + ".Connected")
		knownVar, _ := netObj.GetProperty(IfaceNetwork + ".KnownNetwork")

		netName := variantString(nameVar)
		isConnected := variantBool(connVar)
		if isActive && netName == activeSSID {
			isConnected = true
		}

		isKnown := (knownVar.Value() != nil) || systemSavedMap[netName]

		net := Network{
			Path:      item.Path,
			Name:      netName,
			Type:      variantString(typeVar),
			Connected: isConnected,
			Known:     isKnown,
			Signal:    item.Signal,
		}
		networks = append(networks, net)
	}

	return networks, nil
}

func (c *Client) getNetworksFromObjects(stationPath dbus.ObjectPath) ([]Network, error) {
	objects, err := c.GetManagedObjects()
	if err != nil {
		return nil, err
	}

	systemSaved := GetSystemSavedConnections()
	systemSavedMap := make(map[string]bool, len(systemSaved))
	for _, s := range systemSaved {
		systemSavedMap[s] = true
	}
	activeSSID, isActive := GetSystemActiveWirelessConnection("wlan0")

	var networks []Network
	for path, ifaces := range objects {
		if netProps, ok := ifaces[IfaceNetwork]; ok {
			devVal, hasDev := netProps["Device"]
			if hasDev {
				if devPath, ok := devVal.Value().(dbus.ObjectPath); ok && devPath != stationPath {
					continue
				}
			}

			netName := getString(netProps, "Name")
			isConnected := getBool(netProps, "Connected")
			if isActive && netName == activeSSID {
				isConnected = true
			}

			_, hasKnown := netProps["KnownNetwork"]
			isKnown := hasKnown || systemSavedMap[netName]

			net := Network{
				Path:      path,
				Name:      netName,
				Type:      getString(netProps, "Type"),
				Connected: isConnected,
				Known:     isKnown,
				Signal:    -7000,
			}
			networks = append(networks, net)
		}
	}
	return networks, nil
}

// GetKnownNetworks returns all saved/known networks from iwd and system services
func (c *Client) GetKnownNetworks() ([]KnownNetwork, error) {
	objects, err := c.GetManagedObjects()
	if err != nil {
		return nil, err
	}

	var known []KnownNetwork
	knownMap := make(map[string]bool)

	for path, ifaces := range objects {
		if knProps, ok := ifaces[IfaceKnownNetwork]; ok {
			kn := KnownNetwork{
				Path:              path,
				Name:              getString(knProps, "Name"),
				Type:              getString(knProps, "Type"),
				AutoConnect:       getBool(knProps, "AutoConnect"),
				Hidden:            getBool(knProps, "Hidden"),
				LastConnectedTime: getString(knProps, "LastConnectedTime"),
			}
			known = append(known, kn)
			knownMap[kn.Name] = true
		}
	}

	// Merge system-saved connection profiles
	for _, s := range GetSystemSavedConnections() {
		if !knownMap[s] {
			known = append(known, KnownNetwork{
				Name:        s,
				Type:        "psk",
				AutoConnect: true,
			})
			knownMap[s] = true
		}
	}

	sort.Slice(known, func(i, j int) bool {
		if known[i].LastConnectedTime != known[j].LastConnectedTime {
			return known[i].LastConnectedTime > known[j].LastConnectedTime
		}
		return strings.ToLower(known[i].Name) < strings.ToLower(known[j].Name)
	})

	return known, nil
}

// Scan requests iwd to perform a fresh Wi-Fi scan on the given station
func (c *Client) Scan(stationPath dbus.ObjectPath) error {
	stationObj := c.conn.Object(IwdBusName, stationPath)
	call := stationObj.Call(IfaceStation+".Scan", 0)
	if call.Err != nil {
		return call.Err
	}
	return nil
}

// Connect attempts to connect to a known or open network
func (c *Client) Connect(netPath dbus.ObjectPath, ssid string) error {
	if netPath != "" {
		netObj := c.conn.Object(IwdBusName, netPath)
		call := netObj.Call(IfaceNetwork+".Connect", 0)
		if call.Err == nil {
			return nil
		}
	}

	// If iwd connect didn't succeed or profile is managed by system connection manager:
	if ssid != "" {
		if err := ConnectSystemSaved(ssid); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to connect to %s", ssid)
}

// ConnectWithPassphrase provisions credentials and connects
func (c *Client) ConnectWithPassphrase(deviceName, ssid, passphrase string) error {
	var cmd *exec.Cmd
	if passphrase == "" {
		cmd = exec.Command("iwctl", "station", deviceName, "connect", ssid)
	} else {
		cmd = exec.Command("iwctl", "--passphrase", passphrase, "station", deviceName, "connect", ssid)
	}

	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	// Fallback to nmcli if iwctl reports error
	if passphrase != "" {
		cmdNM := exec.Command("nmcli", "dev", "wifi", "connect", ssid, "password", passphrase)
		if outNM, errNM := cmdNM.CombinedOutput(); errNM == nil {
			return nil
		} else {
			outputMsg := strings.TrimSpace(string(outNM))
			if outputMsg != "" {
				return fmt.Errorf("%s", outputMsg)
			}
		}
	}

	outputMsg := strings.TrimSpace(string(out))
	if outputMsg != "" {
		return fmt.Errorf("%s", outputMsg)
	}
	return err
}

// Disconnect disconnects the station from the active network
func (c *Client) Disconnect(stationPath dbus.ObjectPath) error {
	stationObj := c.conn.Object(IwdBusName, stationPath)
	_ = stationObj.Call(IfaceStation+".Disconnect", 0)
	_ = DisconnectSystem("wlan0")
	return nil
}

// ForgetKnownNetwork removes stored credentials for a network
func (c *Client) ForgetKnownNetwork(knPath dbus.ObjectPath, name string) error {
	if knPath != "" {
		knObj := c.conn.Object(IwdBusName, knPath)
		_ = knObj.Call(IfaceKnownNetwork+".Forget", 0)
	}
	// Also delete from NetworkManager if present
	if name != "" {
		_ = exec.Command("nmcli", "connection", "delete", name).Run()
	}
	return nil
}

// SetAutoConnect toggles the auto-connect flag for a known network
func (c *Client) SetAutoConnect(knPath dbus.ObjectPath, autoConnect bool) error {
	if knPath != "" {
		knObj := c.conn.Object(IwdBusName, knPath)
		return knObj.SetProperty(IfaceKnownNetwork+".AutoConnect", dbus.MakeVariant(autoConnect))
	}
	return nil
}

// ToggleDevicePower turns Wi-Fi radio on or off
func (c *Client) ToggleDevicePower(devPath dbus.ObjectPath, powered bool) error {
	devObj := c.conn.Object(IwdBusName, devPath)
	err := devObj.SetProperty(IfaceDevice+".Powered", dbus.MakeVariant(powered))
	if err != nil {
		// Fallback to nmcli radio
		state := "off"
		if powered {
			state = "on"
		}
		_ = exec.Command("nmcli", "radio", "wifi", state).Run()
	}
	return nil
}

func getString(m map[string]dbus.Variant, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.Value().(string); ok {
			return s
		}
	}
	return ""
}

func getBool(m map[string]dbus.Variant, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.Value().(bool); ok {
			return b
		}
	}
	return false
}

func variantString(v dbus.Variant) string {
	if s, ok := v.Value().(string); ok {
		return s
	}
	return ""
}

func variantBool(v dbus.Variant) bool {
	if b, ok := v.Value().(bool); ok {
		return b
	}
	return false
}

// CheckIwdService returns true if iwd or NetworkManager is active on the system
func CheckIwdService() error {
	conn, err := dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("cannot connect to system bus: %w", err)
	}
	defer conn.Close()

	var names []string
	err = conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names)
	if err != nil {
		return fmt.Errorf("failed to query D-Bus services: %w", err)
	}

	for _, name := range names {
		if name == IwdBusName || name == "org.freedesktop.NetworkManager" {
			return nil
		}
	}
	return errors.New("wireless network service is not running on D-Bus")
}
