package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
	"nm-go/internal/iwd"
)

func fetchDataCmd(client *iwd.Client, background bool) tea.Cmd {
	return func() tea.Msg {
		devices, err := client.GetDevices()
		if err != nil {
			return FetchDataMsg{Err: err}
		}

		if len(devices) == 0 {
			return FetchDataMsg{Devices: devices}
		}

		// Use the first wireless device
		dev := devices[0]
		station, err := client.GetStation(dev.Path)
		if err != nil {
			// Might not be in station mode yet or powered off
			station = &iwd.Station{
				Path:  dev.Path,
				State: "unknown",
			}
		}

		networks, _ := client.GetNetworks(dev.Path)
		knownNetworks, _ := client.GetKnownNetworks()

		return FetchDataMsg{
			Devices:       devices,
			Station:       station,
			Networks:      networks,
			KnownNetworks: knownNetworks,
		}
	}
}

func scanCmd(client *iwd.Client, stationPath dbus.ObjectPath) tea.Cmd {
	return func() tea.Msg {
		err := client.Scan(stationPath)
		// Small delay to allow scan results to populate in iwd
		time.Sleep(1200 * time.Millisecond)
		return ScanFinishedMsg{Err: err}
	}
}

func connectDirectCmd(client *iwd.Client, netPath dbus.ObjectPath, ssid string) tea.Cmd {
	return func() tea.Msg {
		err := client.Connect(netPath)
		return ConnectResultMsg{SSID: ssid, Err: err}
	}
}

func connectWithPassCmd(client *iwd.Client, devName, ssid, passphrase string) tea.Cmd {
	return func() tea.Msg {
		err := client.ConnectWithPassphrase(devName, ssid, passphrase)
		return ConnectResultMsg{SSID: ssid, Err: err}
	}
}

func disconnectCmd(client *iwd.Client, stationPath dbus.ObjectPath) tea.Cmd {
	return func() tea.Msg {
		err := client.Disconnect(stationPath)
		return DisconnectResultMsg{Err: err}
	}
}

func forgetCmd(client *iwd.Client, knPath dbus.ObjectPath, name string) tea.Cmd {
	return func() tea.Msg {
		err := client.ForgetKnownNetwork(knPath)
		return ForgetResultMsg{Name: name, Err: err}
	}
}

func togglePowerCmd(client *iwd.Client, devPath dbus.ObjectPath, powered bool) tea.Cmd {
	return func() tea.Msg {
		err := client.ToggleDevicePower(devPath, powered)
		return PowerResultMsg{Powered: powered, Err: err}
	}
}

func setAutoConnectCmd(client *iwd.Client, knPath dbus.ObjectPath, autoConnect bool) tea.Cmd {
	return func() tea.Msg {
		err := client.SetAutoConnect(knPath, autoConnect)
		return AutoConnectResultMsg{Err: err}
	}
}
