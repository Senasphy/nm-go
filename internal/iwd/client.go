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
	IwdBusName       = "net.connman.iwd"
	ObjectManagerPath = "/"
	InterfaceManager = "org.freedesktop.DBus.ObjectManager"

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
		// Fallback: If GetOrderedNetworks fails or returns empty, inspect managed objects directly
		return c.getNetworksFromObjects(stationPath)
	}

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

		net := Network{
			Path:      item.Path,
			Name:      variantString(nameVar),
			Type:      variantString(typeVar),
			Connected: variantBool(connVar),
			Known:     knownVar.Value() != nil,
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

	var networks []Network
	for path, ifaces := range objects {
		if netProps, ok := ifaces[IfaceNetwork]; ok {
			devVal, hasDev := netProps["Device"]
			if hasDev {
				if devPath, ok := devVal.Value().(dbus.ObjectPath); ok && devPath != stationPath {
					continue
				}
			}

			_, hasKnown := netProps["KnownNetwork"]
			net := Network{
				Path:      path,
				Name:      getString(netProps, "Name"),
				Type:      getString(netProps, "Type"),
				Connected: getBool(netProps, "Connected"),
				Known:     hasKnown,
				Signal:    -7000, // default placeholder if no signal reading
			}
			networks = append(networks, net)
		}
	}
	return networks, nil
}

// GetKnownNetworks returns all saved/known networks
func (c *Client) GetKnownNetworks() ([]KnownNetwork, error) {
	objects, err := c.GetManagedObjects()
	if err != nil {
		return nil, err
	}

	var known []KnownNetwork
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
		}
	}

	sort.Slice(known, func(i, j int) bool {
		// Sort by most recently connected first
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
		// Scan might return OperationNotSupported or Busy if a scan is already running; handle gracefully
		return call.Err
	}
	return nil
}

// Connect attempts to connect to a known or open network via D-Bus
func (c *Client) Connect(netPath dbus.ObjectPath) error {
	netObj := c.conn.Object(IwdBusName, netPath)
	call := netObj.Call(IfaceNetwork+".Connect", 0)
	if call.Err != nil {
		return call.Err
	}
	return nil
}

// ConnectWithPassphrase uses iwctl to connect with a specified passphrase for secured networks
func (c *Client) ConnectWithPassphrase(deviceName, ssid, passphrase string) error {
	var cmd *exec.Cmd
	if passphrase == "" {
		cmd = exec.Command("iwctl", "station", deviceName, "connect", ssid)
	} else {
		cmd = exec.Command("iwctl", "--passphrase", passphrase, "station", deviceName, "connect", ssid)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		outputMsg := strings.TrimSpace(string(out))
		if outputMsg != "" {
			return fmt.Errorf("%s", outputMsg)
		}
		return err
	}
	return nil
}

// Disconnect disconnects the station from the currently connected network
func (c *Client) Disconnect(stationPath dbus.ObjectPath) error {
	stationObj := c.conn.Object(IwdBusName, stationPath)
	call := stationObj.Call(IfaceStation+".Disconnect", 0)
	return call.Err
}

// ForgetKnownNetwork removes stored credentials for a network
func (c *Client) ForgetKnownNetwork(knPath dbus.ObjectPath) error {
	knObj := c.conn.Object(IwdBusName, knPath)
	call := knObj.Call(IfaceKnownNetwork+".Forget", 0)
	return call.Err
}

// SetAutoConnect toggles the auto-connect flag for a known network
func (c *Client) SetAutoConnect(knPath dbus.ObjectPath, autoConnect bool) error {
	knObj := c.conn.Object(IwdBusName, knPath)
	err := knObj.SetProperty(IfaceKnownNetwork+".AutoConnect", dbus.MakeVariant(autoConnect))
	return err
}

// ToggleDevicePower turns Wi-Fi radio on or off for the given device
func (c *Client) ToggleDevicePower(devPath dbus.ObjectPath, powered bool) error {
	devObj := c.conn.Object(IwdBusName, devPath)
	return devObj.SetProperty(IfaceDevice+".Powered", dbus.MakeVariant(powered))
}

// Helper functions for safely reading D-Bus variants
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

// CheckIwdService returns true if iwd is active on the system
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
		if name == IwdBusName {
			return nil
		}
	}
	return errors.New("iwd daemon (net.connman.iwd) is not running on D-Bus")
}
