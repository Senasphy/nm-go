package iwd

import (
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

// Device represents a physical wireless device managed by iwd (e.g., wlan0)
type Device struct {
	Path          dbus.ObjectPath
	Name          string
	Address       string
	Powered       bool
	Mode          string
	Adapter       dbus.ObjectPath
	AdapterName   string
	AdapterModel  string
	AdapterVendor string
}

// Station represents a device operating in station mode
type Station struct {
	Path             dbus.ObjectPath
	State            string // "disconnected", "connected", "connecting", "roaming"
	Scanning         bool
	ConnectedNetwork dbus.ObjectPath
}

// Network represents an available wireless network discovered via scanning
type Network struct {
	Path      dbus.ObjectPath
	Name      string // SSID
	Type      string // "open", "psk", "8021x"
	Connected bool
	Known     bool
	Signal    int16 // signal in centi-dBm (e.g. -6500 = -65 dBm)
}

// SignalDBm returns signal in dBm
func (n Network) SignalDBm() float64 {
	return float64(n.Signal) / 100.0
}

// SignalPercent returns calculated signal strength between 0 and 100%
func (n Network) SignalPercent() int {
	// Wi-Fi signal typically ranges between -100 dBm (unusable) to -40 dBm (maximum)
	dbm := n.SignalDBm()
	if dbm >= -45 {
		return 100
	}
	if dbm <= -95 {
		return 5
	}
	// Linear interpolation between -95 and -45
	pct := int((dbm + 95) * 2)
	if pct < 0 {
		return 0
	}
	if pct > 100 {
		return 100
	}
	return pct
}

// SignalBars returns a visual unicode representation of signal quality
func (n Network) SignalBars() string {
	pct := n.SignalPercent()
	switch {
	case pct >= 75:
		return "▂▄▆█"
	case pct >= 50:
		return "▂▄▆_"
	case pct >= 25:
		return "▂▄__"
	case pct > 0:
		return "▂___"
	default:
		return "____"
	}
}

// SecurityLabel returns a human-readable security description
func (n Network) SecurityLabel() string {
	switch n.Type {
	case "open":
		return "Open"
	case "psk":
		return "WPA2/WPA3 (PSK)"
	case "8021x":
		return "Enterprise (802.1x)"
	default:
		if n.Type == "" {
			return "Unknown"
		}
		return n.Type
	}
}

// KnownNetwork represents a saved network with stored credentials
type KnownNetwork struct {
	Path              dbus.ObjectPath
	Name              string
	Type              string
	AutoConnect       bool
	Hidden            bool
	LastConnectedTime string
	ParsedTime        time.Time
}

// FormattedLastConnected returns a friendly formatted time string
func (kn KnownNetwork) FormattedLastConnected() string {
	if kn.LastConnectedTime == "" {
		return "Never"
	}
	t, err := time.Parse(time.RFC3339, kn.LastConnectedTime)
	if err != nil {
		return kn.LastConnectedTime
	}
	// Format relative or date
	now := time.Now()
	diff := now.Sub(t)
	if diff < 0 {
		return t.Format("Jan 02, 15:04")
	}
	if diff < time.Minute {
		return "Just now"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	}
	if diff < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	}
	return t.Format("Jan 02, 2006")
}
