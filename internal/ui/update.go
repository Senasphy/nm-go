package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"nm-go/internal/iwd"
	"nm-go/internal/theme"
)

type FetchDataMsg struct {
	Devices       []iwd.Device
	Station       *iwd.Station
	Networks      []iwd.Network
	KnownNetworks []iwd.KnownNetwork
	Err           error
}

type PeriodicTickMsg time.Time
type ScanFinishedMsg struct{ Err error }
type ConnectResultMsg struct {
	SSID string
	Err  error
}
type DisconnectResultMsg struct{ Err error }
type ForgetResultMsg struct {
	Name string
	Err  error
}
type PowerResultMsg struct {
	Powered bool
	Err     error
}
type AutoConnectResultMsg struct{ Err error }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case PeriodicTickMsg:
		// Periodic background polling only if not in active modal or connecting
		if m.modal == ModalNone && !m.isConnecting {
			cmds = append(cmds, fetchDataCmd(m.client, true))
		}
		cmds = append(cmds, startPeriodicTick())
		return m, tea.Batch(cmds...)

	case FetchDataMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.setNotification(fmt.Sprintf("IWD Error: %v", msg.Err), NotifError, 5*time.Second)
			return m, nil
		}
		m.devices = msg.Devices
		m.station = msg.Station
		m.networks = msg.Networks
		m.knownNetworks = msg.KnownNetworks
		m.err = nil

		// Identify active network
		m.activeNetwork = nil
		for i := range m.networks {
			if m.networks[i].Connected {
				m.activeNetwork = &m.networks[i]
				break
			}
		}

		// Clamp cursor positions
		filteredNets := m.getFilteredNetworks()
		if len(filteredNets) == 0 {
			m.cursorAvailable = 0
		} else if m.cursorAvailable >= len(filteredNets) {
			m.cursorAvailable = len(filteredNets) - 1
		}

		filteredKnown := m.getFilteredKnownNetworks()
		if len(filteredKnown) == 0 {
			m.cursorSaved = 0
		} else if m.cursorSaved >= len(filteredKnown) {
			m.cursorSaved = len(filteredKnown) - 1
		}
		return m, nil

	case ScanFinishedMsg:
		m.isScanning = false
		m.busyAction = ""
		if msg.Err != nil {
			m.setNotification(fmt.Sprintf("Scan note: %v", msg.Err), NotifWarning, 3*time.Second)
		} else {
			m.setNotification("Scan completed", NotifSuccess, 3*time.Second)
		}
		return m, fetchDataCmd(m.client, false)

	case ConnectResultMsg:
		m.isConnecting = false
		m.busyAction = ""
		if msg.Err != nil {
			m.setNotification(fmt.Sprintf("Failed to connect: %v", msg.Err), NotifError, 6*time.Second)
		} else {
			m.setNotification(fmt.Sprintf("Connected to %s", msg.SSID), NotifSuccess, 4*time.Second)
		}
		return m, fetchDataCmd(m.client, false)

	case DisconnectResultMsg:
		m.busyAction = ""
		if msg.Err != nil {
			m.setNotification(fmt.Sprintf("Disconnect error: %v", msg.Err), NotifError, 4*time.Second)
		} else {
			m.setNotification("Disconnected", NotifInfo, 3*time.Second)
		}
		return m, fetchDataCmd(m.client, false)

	case ForgetResultMsg:
		m.busyAction = ""
		if msg.Err != nil {
			m.setNotification(fmt.Sprintf("Failed to forget: %v", msg.Err), NotifError, 4*time.Second)
		} else {
			m.setNotification(fmt.Sprintf("Forgot network '%s'", msg.Name), NotifSuccess, 3*time.Second)
		}
		return m, fetchDataCmd(m.client, false)

	case PowerResultMsg:
		m.busyAction = ""
		if msg.Err != nil {
			m.setNotification(fmt.Sprintf("Power toggle failed: %v", msg.Err), NotifError, 4*time.Second)
		} else {
			state := "disabled"
			if msg.Powered {
				state = "enabled"
			}
			m.setNotification(fmt.Sprintf("Wi-Fi %s", state), NotifInfo, 3*time.Second)
		}
		return m, fetchDataCmd(m.client, false)

	case AutoConnectResultMsg:
		if msg.Err != nil {
			m.setNotification(fmt.Sprintf("AutoConnect toggle failed: %v", msg.Err), NotifError, 4*time.Second)
		} else {
			m.setNotification("Updated AutoConnect setting", NotifSuccess, 2*time.Second)
		}
		return m, fetchDataCmd(m.client, false)

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle active modal interactions
	if m.modal != ModalNone {
		switch m.modal {
		case ModalThemePicker:
			themes := theme.List()
			switch msg.String() {
			case "esc", "q":
				m.modal = ModalNone
				return m, nil
			case "up", "k":
				if m.themePickerIndex > 0 {
					m.themePickerIndex--
				} else {
					m.themePickerIndex = len(themes) - 1
				}
				// Live preview!
				m.currentTheme = themes[m.themePickerIndex]
				return m, nil
			case "down", "j":
				if m.themePickerIndex < len(themes)-1 {
					m.themePickerIndex++
				} else {
					m.themePickerIndex = 0
				}
				// Live preview!
				m.currentTheme = themes[m.themePickerIndex]
				return m, nil
			case "enter":
				m.currentTheme = themes[m.themePickerIndex]
				m.modal = ModalNone
				m.setNotification(fmt.Sprintf("Theme: %s", m.currentTheme.Name), NotifSuccess, 2*time.Second)
				return m, nil
			}
			return m, nil

		case ModalPassword:
			switch msg.String() {
			case "esc":
				m.modal = ModalNone
				return m, nil
			case "ctrl+r":
				m.passwordModal.ToggleEcho()
				return m, nil
			case "enter":
				pass := m.passwordModal.Input.Value()
				targetNet := m.passwordModal.Network
				m.modal = ModalNone
				m.isConnecting = true
				m.busyAction = fmt.Sprintf("Connecting to %s...", targetNet.Name)

				devName := "wlan0"
				if len(m.devices) > 0 && m.selectedDev < len(m.devices) {
					devName = m.devices[m.selectedDev].Name
				}
				return m, connectWithPassCmd(m.client, devName, targetNet.Name, pass)
			default:
				var cmd tea.Cmd
				m.passwordModal.Input, cmd = m.passwordModal.Input.Update(msg)
				return m, cmd
			}

		case ModalConfirmForget:
			switch msg.String() {
			case "y", "Y", "enter":
				target := m.targetForget
				m.modal = ModalNone
				m.targetForget = nil
				if target != nil {
					m.busyAction = fmt.Sprintf("Forgetting %s...", target.Name)
					return m, forgetCmd(m.client, target.Path, target.Name)
				}
				return m, nil
			case "n", "N", "esc":
				m.modal = ModalNone
				m.targetForget = nil
				return m, nil
			}

		case ModalHelp:
			if msg.String() == "esc" || msg.String() == "?" || msg.String() == "q" || msg.String() == "enter" {
				m.modal = ModalNone
				return m, nil
			}
		}
		return m, nil
	}

	// Handle search/filtering input mode
	if m.filtering {
		switch msg.String() {
		case "esc", "ctrl+c":
			m.filtering = false
			m.filterInput.SetValue("")
			return m, nil
		case "enter", "down", "tab":
			m.filtering = false
			return m, nil
		default:
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)
			return m, cmd
		}
	}

	// Normal keybindings
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Help):
		m.modal = ModalHelp
		return m, nil

	case key.Matches(msg, keys.Theme):
		themes := theme.List()
		for i, th := range themes {
			if th.ID == m.currentTheme.ID {
				m.themePickerIndex = i
				break
			}
		}
		m.modal = ModalThemePicker
		return m, nil

	case key.Matches(msg, keys.Tab):
		m.activeTab = (m.activeTab + 1) % 3
		return m, nil

	case key.Matches(msg, keys.BackTab):
		if m.activeTab == 0 {
			m.activeTab = 2
		} else {
			m.activeTab--
		}
		return m, nil

	case key.Matches(msg, keys.Filter):
		m.filtering = true
		m.filterInput.Focus()
		return m, nil

	case key.Matches(msg, keys.Up):
		m.navigateUp()
		return m, nil

	case key.Matches(msg, keys.Down):
		m.navigateDown()
		return m, nil

	case key.Matches(msg, keys.Scan):
		if len(m.devices) > 0 && m.selectedDev < len(m.devices) {
			m.isScanning = true
			m.busyAction = "Scanning for networks..."
			return m, scanCmd(m.client, m.devices[m.selectedDev].Path)
		}

	case key.Matches(msg, keys.Connect):
		return m.handleConnect()

	case key.Matches(msg, keys.Disconnect):
		if len(m.devices) > 0 && m.selectedDev < len(m.devices) {
			m.busyAction = "Disconnecting..."
			return m, disconnectCmd(m.client, m.devices[m.selectedDev].Path)
		}

	case key.Matches(msg, keys.Forget):
		if m.activeTab == TabSaved {
			known := m.getFilteredKnownNetworks()
			if len(known) > 0 && m.cursorSaved < len(known) {
				m.targetForget = &known[m.cursorSaved]
				m.modal = ModalConfirmForget
			}
		} else if m.activeTab == TabAvailable {
			avail := m.getFilteredNetworks()
			if len(avail) > 0 && m.cursorAvailable < len(avail) && avail[m.cursorAvailable].Known {
				for i := range m.knownNetworks {
					if m.knownNetworks[i].Name == avail[m.cursorAvailable].Name {
						m.targetForget = &m.knownNetworks[i]
						m.modal = ModalConfirmForget
						break
					}
				}
			}
		}
		return m, nil

	case key.Matches(msg, keys.AutoConn):
		if m.activeTab == TabSaved {
			known := m.getFilteredKnownNetworks()
			if len(known) > 0 && m.cursorSaved < len(known) {
				target := known[m.cursorSaved]
				return m, setAutoConnectCmd(m.client, target.Path, !target.AutoConnect)
			}
		}

	case key.Matches(msg, keys.Power):
		if len(m.devices) > 0 && m.selectedDev < len(m.devices) {
			dev := m.devices[m.selectedDev]
			m.busyAction = "Toggling Wi-Fi..."
			return m, togglePowerCmd(m.client, dev.Path, !dev.Powered)
		}
	}

	return m, nil
}

func (m *Model) navigateUp() {
	switch m.activeTab {
	case TabAvailable:
		if m.cursorAvailable > 0 {
			m.cursorAvailable--
		}
	case TabSaved:
		if m.cursorSaved > 0 {
			m.cursorSaved--
		}
	}
}

func (m *Model) navigateDown() {
	switch m.activeTab {
	case TabAvailable:
		filtered := m.getFilteredNetworks()
		if m.cursorAvailable < len(filtered)-1 {
			m.cursorAvailable++
		}
	case TabSaved:
		filtered := m.getFilteredKnownNetworks()
		if m.cursorSaved < len(filtered)-1 {
			m.cursorSaved++
		}
	}
}

func (m *Model) handleConnect() (tea.Model, tea.Cmd) {
	if m.activeTab == TabAvailable {
		filtered := m.getFilteredNetworks()
		if len(filtered) == 0 || m.cursorAvailable >= len(filtered) {
			return m, nil
		}
		selected := filtered[m.cursorAvailable]

		if selected.Connected {
			m.setNotification(fmt.Sprintf("Already connected to %s", selected.Name), NotifInfo, 3*time.Second)
			return m, nil
		}

		if selected.Known || selected.Type == "open" {
			m.isConnecting = true
			m.busyAction = fmt.Sprintf("Connecting to %s...", selected.Name)
			return m, connectDirectCmd(m.client, selected.Path, selected.Name)
		}

		m.passwordModal = NewPasswordModal(selected, m.currentTheme)
		m.modal = ModalPassword
		return m, nil
	}

	if m.activeTab == TabSaved {
		filtered := m.getFilteredKnownNetworks()
		if len(filtered) == 0 || m.cursorSaved >= len(filtered) {
			return m, nil
		}
		selectedKnown := filtered[m.cursorSaved]

		for _, net := range m.networks {
			if net.Name == selectedKnown.Name {
				if net.Connected {
					m.setNotification(fmt.Sprintf("Already connected to %s", net.Name), NotifInfo, 3*time.Second)
					return m, nil
				}
				m.isConnecting = true
				m.busyAction = fmt.Sprintf("Connecting to %s...", net.Name)
				return m, connectDirectCmd(m.client, net.Path, net.Name)
			}
		}

		devName := "wlan0"
		if len(m.devices) > 0 && m.selectedDev < len(m.devices) {
			devName = m.devices[m.selectedDev].Name
		}
		m.isConnecting = true
		m.busyAction = fmt.Sprintf("Connecting to %s...", selectedKnown.Name)
		return m, connectWithPassCmd(m.client, devName, selectedKnown.Name, "")
	}

	return m, nil
}

func (m *Model) setNotification(msg string, t NotificationType, duration time.Duration) {
	m.notification = &Notification{
		Message: msg,
		Type:    t,
		Expires: time.Now().Add(duration),
	}
}

func (m Model) getFilteredNetworks() []iwd.Network {
	filter := strings.TrimSpace(strings.ToLower(m.filterInput.Value()))
	if filter == "" {
		return m.networks
	}
	var res []iwd.Network
	for _, n := range m.networks {
		if strings.Contains(strings.ToLower(n.Name), filter) {
			res = append(res, n)
		}
	}
	return res
}

func (m Model) getFilteredKnownNetworks() []iwd.KnownNetwork {
	filter := strings.TrimSpace(strings.ToLower(m.filterInput.Value()))
	if filter == "" {
		return m.knownNetworks
	}
	var res []iwd.KnownNetwork
	for _, kn := range m.knownNetworks {
		if strings.Contains(strings.ToLower(kn.Name), filter) {
			res = append(res, kn)
		}
	}
	return res
}
