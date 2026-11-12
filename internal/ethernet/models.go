package ethernet

import (
	"fmt"
	"strings"
)

// Device represents an Ethernet (wired) network interface.
type Device struct {
	Name        string   // Interface name (e.g. "enp0s31f6", "eth0")
	MAC         string   // Hardware MAC address (e.g. "c4:65:16:0f:4f:67")
	State       string   // High-level state ("connected", "disconnected", "unplugged", "connecting", "unavailable", "down")
	Carrier     bool     // Physical link carrier detected (cable plugged in and link up)
	OperState   string   // Operational state ("up", "down", "unknown", "dormant", "lowerlayerdown")
	Speed       int      // Link speed in Mbits/sec (e.g. 100, 1000, 2500, 10000; -1 or 0 if unknown/down)
	Duplex      string   // "full", "half", or "unknown"
	IP4         []string // Assigned IPv4 addresses (with CIDR, e.g. "192.168.1.50/24")
	IP6         []string // Assigned IPv6 addresses
	Gateway     string   // Default gateway (IPv4 or IPv6)
	MTU         int      // Maximum transmission unit (e.g. 1500)
	Driver      string   // Kernel driver name (e.g. "e1000e", "r8169")
	Connection  string   // Active NetworkManager connection profile name (e.g. "Wired connection 1")
	IsPhysical  bool     // True if hardware-backed (PCI/USB/on-board physical NIC)
	IsUp        bool     // Administrative status (net.FlagUp)
	RxBytes     uint64   // Total received bytes
	TxBytes     uint64   // Total transmitted bytes
}

// PrimaryIP returns the first IPv4 address, or the first IPv6 address if no IPv4 is available.
// If no address is configured, it returns "-".
func (d Device) PrimaryIP() string {
	if len(d.IP4) > 0 {
		return d.IP4[0]
	}
	if len(d.IP6) > 0 {
		return d.IP6[0]
	}
	return "-"
}

// PrimaryIPClean returns the primary IP address without the CIDR prefix length.
func (d Device) PrimaryIPClean() string {
	ip := d.PrimaryIP()
	if ip == "-" {
		return "-"
	}
	if idx := strings.Index(ip, "/"); idx != -1 {
		return ip[:idx]
	}
	return ip
}

// FormattedSpeed returns a human-friendly representation of link speed and duplex mode.
func (d Device) FormattedSpeed() string {
	if !d.Carrier || d.Speed <= 0 {
		return "-"
	}

	var speedStr string
	if d.Speed >= 1000 {
		if d.Speed%1000 == 0 {
			speedStr = fmt.Sprintf("%d Gbps", d.Speed/1000)
		} else {
			speedStr = fmt.Sprintf("%.1f Gbps", float64(d.Speed)/1000.0)
		}
	} else {
		speedStr = fmt.Sprintf("%d Mbps", d.Speed)
	}

	switch strings.ToLower(d.Duplex) {
	case "full":
		return speedStr + " (Full)"
	case "half":
		return speedStr + " (Half)"
	default:
		return speedStr
	}
}

// IsConnected returns true if the interface is actively connected with a carrier and an assigned IP,
// or marked connected by the network manager.
func (d Device) IsConnected() bool {
	if strings.EqualFold(d.State, "connected") {
		return true
	}
	return d.Carrier && len(d.IP4) > 0
}

// StatusLabel returns a user-friendly status description.
func (d Device) StatusLabel() string {
	if d.IsConnected() {
		return "Connected"
	}
	if strings.EqualFold(d.State, "connecting") {
		return "Connecting"
	}
	if d.Carrier {
		return "Link Up (No IP)"
	}
	if strings.EqualFold(d.State, "unavailable") {
		return "Cable Unplugged"
	}
	if strings.EqualFold(d.OperState, "down") || (!d.IsUp && d.OperState != "") {
		return "Disabled / Down"
	}
	return "Unplugged"
}
