package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
	"github.com/Senasphy/nm-go/internal/ethernet"
	"github.com/Senasphy/nm-go/internal/iwd"
)

func fetchDataCmd(client *iwd.Client, ethClient *ethernet.Client, background bool) tea.Cmd {
	return func() tea.Msg {
		var ethDevices []ethernet.Device
		if ethClient != nil {
			ethDevices, _ = ethClient.GetDevices()
		}

		if client == nil {
			return FetchDataMsg{
				EthDevices: ethDevices,
			}
		}

		devices, err := client.GetDevices()
		if err != nil {
			return FetchDataMsg{
				EthDevices: ethDevices,
				Err:        err,
			}
		}

		if len(devices) == 0 {
			return FetchDataMsg{
				Devices:    devices,
				EthDevices: ethDevices,
			}
		}

		dev := devices[0]
		station, err := client.GetStation(dev.Path)
		if err != nil {
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
			EthDevices:    ethDevices,
		}
	}
}

func connectEthCmd(ethClient *ethernet.Client, devName string) tea.Cmd {
	return func() tea.Msg {
		var err error
		if ethClient != nil {
			err = ethClient.Connect(devName)
		} else {
			err = fmt.Errorf("ethernet client not initialized")
		}
		return ConnectEthResultMsg{DevName: devName, Err: err}
	}
}

func disconnectEthCmd(ethClient *ethernet.Client, devName string) tea.Cmd {
	return func() tea.Msg {
		var err error
		if ethClient != nil {
			err = ethClient.Disconnect(devName)
		} else {
			err = fmt.Errorf("ethernet client not initialized")
		}
		return DisconnectEthResultMsg{DevName: devName, Err: err}
	}
}

func scanCmd(client *iwd.Client, stationPath dbus.ObjectPath) tea.Cmd {
	return func() tea.Msg {
		err := client.Scan(stationPath)
		time.Sleep(1200 * time.Millisecond)
		return ScanFinishedMsg{Err: err}
	}
}

func connectDirectCmd(client *iwd.Client, netPath dbus.ObjectPath, ssid string) tea.Cmd {
	return func() tea.Msg {
		err := client.Connect(netPath, ssid)
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
		err := client.ForgetKnownNetwork(knPath, name)
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
