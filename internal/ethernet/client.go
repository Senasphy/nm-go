package ethernet

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var (
	// Default sysfs and procfs locations (can be overridden in tests)
	sysfsNetDir   = "/sys/class/net"
	procRouteFile = "/proc/net/route"
)

// Client handles discovery and management of Ethernet devices.
type Client struct{}

// NewClient returns a new Ethernet client instance.
func NewClient() *Client {
	return &Client{}
}

// GetDevices returns all discovered Ethernet network devices.
func (c *Client) GetDevices() ([]Device, error) {
	return getDevicesFromSysfs(sysfsNetDir)
}

// Connect attempts to bring up or connect an Ethernet interface.
func (c *Client) Connect(devName string) error {
	if strings.TrimSpace(devName) == "" {
		return errors.New("device name cannot be empty")
	}

	// 1. Try NetworkManager CLI if available
	if hasCommand("nmcli") {
		cmd := exec.Command("nmcli", "device", "connect", devName)
		out, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		// If direct device connect failed, try connection up for any matching profile
		profiles := getProfilesForDevice(devName)
		for _, prof := range profiles {
			cmdUp := exec.Command("nmcli", "connection", "up", prof)
			if outUp, errUp := cmdUp.CombinedOutput(); errUp == nil {
				return nil
			} else if strings.TrimSpace(string(outUp)) != "" {
				// Continue trying or save error
			}
		}
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
	}

	// 2. Fallback to ip link set <dev> up
	if hasCommand("ip") {
		cmd := exec.Command("ip", "link", "set", devName, "up")
		out, err := cmd.CombinedOutput()
		if err != nil {
			msg := strings.TrimSpace(string(out))
			if msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return err
		}
		return nil
	}

	return fmt.Errorf("no suitable tool found to activate %s (install nmcli or iproute2)", devName)
}

// Disconnect attempts to disconnect or bring down an Ethernet interface.
func (c *Client) Disconnect(devName string) error {
	if strings.TrimSpace(devName) == "" {
		return errors.New("device name cannot be empty")
	}

	// 1. Try NetworkManager CLI if available
	if hasCommand("nmcli") {
		cmd := exec.Command("nmcli", "device", "disconnect", devName)
		out, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
	}

	// 2. Fallback to ip link set <dev> down
	if hasCommand("ip") {
		cmd := exec.Command("ip", "link", "set", devName, "down")
		out, err := cmd.CombinedOutput()
		if err != nil {
			msg := strings.TrimSpace(string(out))
			if msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return err
		}
		return nil
	}

	return fmt.Errorf("no suitable tool found to deactivate %s", devName)
}

// ToggleDevice turns the device link up or down.
func (c *Client) ToggleDevice(devName string, up bool) error {
	if up {
		return c.Connect(devName)
	}
	return c.Disconnect(devName)
}

// getDevicesFromSysfs scans the given sysfs directory for ethernet devices.
func getDevicesFromSysfs(baseDir string) ([]Device, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read network devices from %s: %w", baseDir, err)
	}

	nmStatus := getNMDeviceStatus()
	gateways := getDefaultGateways(procRouteFile)

	var devices []Device

	for _, entry := range entries {
		devName := entry.Name()
		devPath := filepath.Join(baseDir, devName)

		if !isEthernetDevice(devPath, devName) {
			continue
		}

		dev := buildDevice(devPath, devName, nmStatus, gateways)
		devices = append(devices, dev)
	}

	// Sort devices: Connected first, then physical devices by name, then others
	sort.Slice(devices, func(i, j int) bool {
		if devices[i].IsConnected() != devices[j].IsConnected() {
			return devices[i].IsConnected()
		}
		if devices[i].IsPhysical != devices[j].IsPhysical {
			return devices[i].IsPhysical
		}
		return devices[i].Name < devices[j].Name
	})

	return devices, nil
}

// isEthernetDevice checks if a sysfs network directory represents a wired Ethernet device.
func isEthernetDevice(devPath, devName string) bool {
	// 1. Exclude loopback
	if devName == "lo" {
		return false
	}

	// 2. Exclude wireless interfaces (has 'wireless' directory or 'phy80211')
	if fileExists(filepath.Join(devPath, "wireless")) || fileExists(filepath.Join(devPath, "phy80211")) {
		return false
	}

	// Wireless interface naming conventions
	if strings.HasPrefix(devName, "wl") || strings.HasPrefix(devName, "wlan") || strings.HasPrefix(devName, "wifi") {
		return false
	}

	// 3. Exclude common virtual container/bridge interfaces unless they are physical or virtual ethernet
	if strings.HasPrefix(devName, "docker") || strings.HasPrefix(devName, "br-") || strings.HasPrefix(devName, "virbr") {
		return false
	}

	// 4. Check device type: Ethernet is ARPHRD_ETHER (type 1)
	typeBytes, err := os.ReadFile(filepath.Join(devPath, "type"))
	if err == nil {
		devType := strings.TrimSpace(string(typeBytes))
		if devType != "1" {
			return false
		}
	}

	// 5. Must either have physical device backing OR start with standard ethernet prefixes
	hasDeviceLink := fileExists(filepath.Join(devPath, "device"))
	isEthName := strings.HasPrefix(devName, "en") || strings.HasPrefix(devName, "eth")

	return hasDeviceLink || isEthName
}

func buildDevice(devPath, devName string, nmStatus map[string]nmDevInfo, gateways map[string]string) Device {
	dev := Device{
		Name:       devName,
		IsPhysical: fileExists(filepath.Join(devPath, "device")),
		Duplex:     "unknown",
		Speed:      -1,
	}

	// 1. MAC address
	if addrBytes, err := os.ReadFile(filepath.Join(devPath, "address")); err == nil {
		dev.MAC = strings.TrimSpace(string(addrBytes))
	}

	// 2. Operstate
	if operBytes, err := os.ReadFile(filepath.Join(devPath, "operstate")); err == nil {
		dev.OperState = strings.TrimSpace(string(operBytes))
	}

	// 3. Carrier (1 = link up, 0 = link down/unplugged)
	if carrierBytes, err := os.ReadFile(filepath.Join(devPath, "carrier")); err == nil {
		dev.Carrier = strings.TrimSpace(string(carrierBytes)) == "1"
	}

	// 4. MTU
	if mtuBytes, err := os.ReadFile(filepath.Join(devPath, "mtu")); err == nil {
		if mtuVal, err := strconv.Atoi(strings.TrimSpace(string(mtuBytes))); err == nil {
			dev.MTU = mtuVal
		}
	}

	// 5. Speed (only valid if carrier is active)
	if dev.Carrier {
		if speedBytes, err := os.ReadFile(filepath.Join(devPath, "speed")); err == nil {
			if spdVal, err := strconv.Atoi(strings.TrimSpace(string(speedBytes))); err == nil && spdVal > 0 {
				dev.Speed = spdVal
			}
		}
	}

	// 6. Duplex
	if dev.Carrier {
		if duplexBytes, err := os.ReadFile(filepath.Join(devPath, "duplex")); err == nil {
			dev.Duplex = strings.TrimSpace(string(duplexBytes))
		}
	}

	// 7. Driver
	driverPath := filepath.Join(devPath, "device", "driver")
	if target, err := filepath.EvalSymlinks(driverPath); err == nil {
		dev.Driver = filepath.Base(target)
	}

	// 8. Statistics
	if rxBytes, err := os.ReadFile(filepath.Join(devPath, "statistics", "rx_bytes")); err == nil {
		if val, err := strconv.ParseUint(strings.TrimSpace(string(rxBytes)), 10, 64); err == nil {
			dev.RxBytes = val
		}
	}
	if txBytes, err := os.ReadFile(filepath.Join(devPath, "statistics", "tx_bytes")); err == nil {
		if val, err := strconv.ParseUint(strings.TrimSpace(string(txBytes)), 10, 64); err == nil {
			dev.TxBytes = val
		}
	}

	// 9. IP Addresses and Link Up flag via net.Interface
	if iface, err := net.InterfaceByName(devName); err == nil {
		dev.IsUp = (iface.Flags & net.FlagUp) != 0
		if dev.MAC == "" && len(iface.HardwareAddr) > 0 {
			dev.MAC = iface.HardwareAddr.String()
		}
		if dev.MTU == 0 {
			dev.MTU = iface.MTU
		}

		if addrs, err := iface.Addrs(); err == nil {
			for _, addr := range addrs {
				ipStr := addr.String()
				if strings.Contains(ipStr, ":") {
					dev.IP6 = append(dev.IP6, ipStr)
				} else {
					dev.IP4 = append(dev.IP4, ipStr)
				}
			}
		}
	}

	// 10. Gateway
	if gw, ok := gateways[devName]; ok {
		dev.Gateway = gw
	}

	// 11. Cross-reference NetworkManager device status
	if nmInfo, ok := nmStatus[devName]; ok {
		dev.State = nmInfo.State
		dev.Connection = nmInfo.Connection
	} else {
		// Infer state from carrier and operstate
		if dev.Carrier && len(dev.IP4) > 0 {
			dev.State = "connected"
		} else if dev.Carrier {
			dev.State = "connecting"
		} else if !dev.IsUp || strings.EqualFold(dev.OperState, "down") {
			dev.State = "down"
		} else {
			dev.State = "unplugged"
		}
	}

	return dev
}

type nmDevInfo struct {
	State      string
	Connection string
}

// getNMDeviceStatus parses nmcli device status if available.
func getNMDeviceStatus() map[string]nmDevInfo {
	result := make(map[string]nmDevInfo)
	if !hasCommand("nmcli") {
		return result
	}

	cmd := exec.Command("nmcli", "-t", "-f", "DEVICE,TYPE,STATE,CONNECTION", "device", "status")
	out, err := cmd.Output()
	if err != nil {
		return result
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			dev := parts[0]
			state := parts[2]
			conn := ""
			if len(parts) >= 4 {
				conn = parts[3]
				if conn == "--" {
					conn = ""
				}
			}
			result[dev] = nmDevInfo{
				State:      state,
				Connection: conn,
			}
		}
	}

	return result
}

// getProfilesForDevice finds all NetworkManager connection profile names for a device.
func getProfilesForDevice(devName string) []string {
	var profiles []string
	if !hasCommand("nmcli") {
		return profiles
	}

	cmd := exec.Command("nmcli", "-t", "-f", "NAME,TYPE,DEVICE", "connection", "show")
	out, err := cmd.Output()
	if err != nil {
		return profiles
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			name := parts[0]
			connType := parts[1]
			dev := parts[2]
			if strings.Contains(connType, "ethernet") || strings.Contains(connType, "802-3") {
				if dev == devName || dev == "" {
					profiles = append(profiles, name)
				}
			}
		}
	}
	return profiles
}

// getDefaultGateways parses /proc/net/route for default IPv4 gateways per interface.
func getDefaultGateways(routePath string) map[string]string {
	gateways := make(map[string]string)
	data, err := os.ReadFile(routePath)
	if err != nil {
		return gateways
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			iface := fields[0]
			dest := fields[1]
			gwHex := fields[2]

			// Destination 00000000 indicates a default gateway route
			if dest == "00000000" && gwHex != "00000000" {
				if gwIP := parseHexIPv4(gwHex); gwIP != "" {
					if _, exists := gateways[iface]; !exists {
						gateways[iface] = gwIP
					}
				}
			}
		}
	}

	return gateways
}

// parseHexIPv4 converts an 8-character little-endian hex string from /proc/net/route into an IPv4 string.
func parseHexIPv4(hexStr string) string {
	if len(hexStr) != 8 {
		return ""
	}
	bytes, err := hex.DecodeString(hexStr)
	if err != nil || len(bytes) != 4 {
		return ""
	}
	// /proc/net/route stores IPv4 in network byte order in hex:
	// E.g. "019B1C0A": bytes[0]=0x01, bytes[1]=0x9B, bytes[2]=0x1C, bytes[3]=0x0A
	// In little-endian format from kernel: byte 0 is lowest octet (1), byte 3 is highest (10) -> 10.28.155.1
	ip := net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
	return ip.String()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
