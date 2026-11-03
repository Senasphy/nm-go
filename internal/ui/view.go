package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing nm-go..."
	}

	// If modal is active, overlay it
	if m.modal != ModalNone {
		switch m.modal {
		case ModalPassword:
			return renderPasswordModal(m.passwordModal, m.width)
		case ModalConfirmForget:
			name := "selected network"
			if m.targetForget != nil {
				name = m.targetForget.Name
			}
			return renderConfirmForgetModal(name, m.width)
		case ModalHelp:
			return renderHelpModal(m.width)
		}
	}

	header := m.renderHeader()
	tabs := m.renderTabs()
	content := m.renderContent()
	footer := m.renderFooter()

	mainView := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		content,
		footer,
	)

	return mainView
}

func (m Model) renderHeader() string {
	logo := StyleHeader.Render(" 󰤨  nm-go ")

	var devInfo string
	if len(m.devices) > 0 && m.selectedDev < len(m.devices) {
		dev := m.devices[m.selectedDev]
		powerBadge := StyleBadgeSuccess.Render("POWER: ON")
		if !dev.Powered {
			powerBadge = StyleBadgeDanger.Render("POWER: OFF")
		}
		devInfo = lipgloss.JoinHorizontal(lipgloss.Center,
			lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(fmt.Sprintf(" dev:%s ", dev.Name)),
			powerBadge,
		)
	}

	var stateBadge string
	if m.isScanning {
		stateBadge = StyleBadgeWarning.Render("SCANNING " + m.spinner.View())
	} else if m.isConnecting {
		stateBadge = StyleBadgeWarning.Render("CONNECTING " + m.spinner.View())
	} else if m.station != nil {
		switch m.station.State {
		case "connected":
			ssid := "Wi-Fi"
			if m.activeNetwork != nil {
				ssid = m.activeNetwork.Name
			}
			stateBadge = StyleBadgeSuccess.Render("CONNECTED: " + ssid)
		case "connecting":
			stateBadge = StyleBadgeWarning.Render("CONNECTING")
		case "roaming":
			stateBadge = StyleBadgeWarning.Render("ROAMING")
		default:
			stateBadge = StyleBadgeMuted.Render("DISCONNECTED")
		}
	} else {
		stateBadge = StyleBadgeMuted.Render("DISCONNECTED")
	}

	rightInfo := ""
	if m.activeNetwork != nil {
		bars := m.activeNetwork.SignalBars()
		pct := m.activeNetwork.SignalPercent()
		sigStyled := SignalStyle(pct).Render(fmt.Sprintf("%s %d%%", bars, pct))
		rightInfo = sigStyled
	}

	leftHeader := lipgloss.JoinHorizontal(lipgloss.Center, logo, "  ", devInfo, stateBadge)
	gap := m.width - lipgloss.Width(leftHeader) - lipgloss.Width(rightInfo) - 2
	if gap < 1 {
		gap = 1
	}

	headerLine := lipgloss.JoinHorizontal(lipgloss.Center, leftHeader, strings.Repeat(" ", gap), rightInfo)
	return lipgloss.NewStyle().MarginBottom(1).Render(headerLine)
}

func (m Model) renderTabs() string {
	availCount := len(m.networks)
	savedCount := len(m.knownNetworks)

	tab0Text := fmt.Sprintf("1. Available (%d)", availCount)
	tab1Text := fmt.Sprintf("2. Saved (%d)", savedCount)
	tab2Text := "3. Device Info"

	var t0, t1, t2 string
	if m.activeTab == TabAvailable {
		t0 = StyleTabActive.Render(tab0Text)
	} else {
		t0 = StyleTabInactive.Render(tab0Text)
	}

	if m.activeTab == TabSaved {
		t1 = StyleTabActive.Render(tab1Text)
	} else {
		t1 = StyleTabInactive.Render(tab1Text)
	}

	if m.activeTab == TabDetails {
		t2 = StyleTabActive.Render(tab2Text)
	} else {
		t2 = StyleTabInactive.Render(tab2Text)
	}

	tabsRow := lipgloss.JoinHorizontal(lipgloss.Top, t0, t1, t2)

	// Filter box display on the right
	var filterDisplay string
	if m.filtering {
		filterDisplay = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1).
			Render("Filter: " + m.filterInput.View())
	} else if m.filterInput.Value() != "" {
		filterDisplay = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Render(fmt.Sprintf("Filtered by '%s' ([Esc] to clear)", m.filterInput.Value()))
	}

	gap := m.width - lipgloss.Width(tabsRow) - lipgloss.Width(filterDisplay) - 2
	if gap < 1 {
		gap = 1
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, tabsRow, strings.Repeat(" ", gap), filterDisplay)
}

func (m Model) renderContent() string {
	switch m.activeTab {
	case TabAvailable:
		return m.renderAvailableNetworks()
	case TabSaved:
		return m.renderSavedNetworks()
	case TabDetails:
		return m.renderDeviceDetails()
	default:
		return ""
	}
}

func (m Model) renderAvailableNetworks() string {
	networks := m.getFilteredNetworks()

	colWidthName := 32
	colWidthSec := 20
	colWidthSignal := 18

	header := StyleTableHeader.Render(
		fmt.Sprintf("  %-32s %-20s %-18s %s", "SSID", "SECURITY", "SIGNAL", "STATUS"),
	)

	separator := lipgloss.NewStyle().Foreground(ColorMuted).Render(
		"  " + strings.Repeat("─", colWidthName) + " " + strings.Repeat("─", colWidthSec) + " " + strings.Repeat("─", colWidthSignal) + " " + strings.Repeat("─", 16),
	)

	if len(networks) == 0 {
		emptyMsg := "No networks found."
		if m.filterInput.Value() != "" {
			emptyMsg = fmt.Sprintf("No networks matching '%s'.", m.filterInput.Value())
		} else {
			emptyMsg = "No networks visible. Press [s] or [r] to trigger a scan."
		}
		emptyBlock := lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(2, 4).
			Render(emptyMsg)
		return lipgloss.JoinVertical(lipgloss.Left, header, separator, emptyBlock)
	}

	// Calculate visible window
	maxVisible := m.height - 10
	if maxVisible < 5 {
		maxVisible = 5
	}

	start := 0
	if m.cursorAvailable >= maxVisible {
		start = m.cursorAvailable - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(networks) {
		end = len(networks)
	}

	var rows []string
	rows = append(rows, header, separator)

	for i := start; i < end; i++ {
		net := networks[i]
		isSelected := i == m.cursorAvailable

		cursorMarker := "  "
		if isSelected {
			cursorMarker = "▶ "
		}

		nameStr := net.Name
		if len(nameStr) > colWidthName-1 {
			nameStr = nameStr[:colWidthName-4] + "..."
		}

		secStr := net.SecurityLabel()
		if len(secStr) > colWidthSec-1 {
			secStr = secStr[:colWidthSec-4] + "..."
		}

		sigPct := net.SignalPercent()
		sigBars := net.SignalBars()
		sigStr := fmt.Sprintf("%s %3d%% (%4.1f dBm)", sigBars, sigPct, net.SignalDBm())

		statusTag := ""
		if net.Connected {
			statusTag = StyleConnectedTag.Render("✔ Connected")
		} else if net.Known {
			statusTag = StyleKnownTag.Render("★ Saved")
		} else if net.Type == "open" {
			statusTag = lipgloss.NewStyle().Foreground(ColorWarning).Render("● Open")
		}

		rowContent := fmt.Sprintf("%s%-32s %-20s %-22s %s",
			cursorMarker,
			nameStr,
			secStr,
			SignalStyle(sigPct).Render(sigStr),
			statusTag,
		)

		if isSelected {
			rows = append(rows, StyleRowSelected.Width(m.width-2).Render(rowContent))
		} else {
			rows = append(rows, StyleRow.Width(m.width-2).Render(rowContent))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderSavedNetworks() string {
	known := m.getFilteredKnownNetworks()

	colWidthName := 32
	colWidthSec := 18
	colWidthAuto := 16

	header := StyleTableHeader.Render(
		fmt.Sprintf("  %-32s %-18s %-16s %s", "SSID", "SECURITY", "AUTO-CONNECT", "LAST CONNECTED"),
	)

	separator := lipgloss.NewStyle().Foreground(ColorMuted).Render(
		"  " + strings.Repeat("─", colWidthName) + " " + strings.Repeat("─", colWidthSec) + " " + strings.Repeat("─", colWidthAuto) + " " + strings.Repeat("─", 20),
	)

	if len(known) == 0 {
		emptyMsg := "No saved networks."
		if m.filterInput.Value() != "" {
			emptyMsg = fmt.Sprintf("No saved networks matching '%s'.", m.filterInput.Value())
		}
		emptyBlock := lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(2, 4).
			Render(emptyMsg)
		return lipgloss.JoinVertical(lipgloss.Left, header, separator, emptyBlock)
	}

	maxVisible := m.height - 10
	if maxVisible < 5 {
		maxVisible = 5
	}

	start := 0
	if m.cursorSaved >= maxVisible {
		start = m.cursorSaved - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(known) {
		end = len(known)
	}

	var rows []string
	rows = append(rows, header, separator)

	for i := start; i < end; i++ {
		kn := known[i]
		isSelected := i == m.cursorSaved

		cursorMarker := "  "
		if isSelected {
			cursorMarker = "▶ "
		}

		nameStr := kn.Name
		if len(nameStr) > colWidthName-1 {
			nameStr = nameStr[:colWidthName-4] + "..."
		}

		autoStr := "No"
		if kn.AutoConnect {
			autoStr = StyleConnectedTag.Render("✔ Yes")
		} else {
			autoStr = lipgloss.NewStyle().Foreground(ColorMuted).Render("✘ No")
		}

		lastStr := kn.FormattedLastConnected()

		rowContent := fmt.Sprintf("%s%-32s %-18s %-24s %s",
			cursorMarker,
			nameStr,
			kn.Type,
			autoStr,
			lastStr,
		)

		if isSelected {
			rows = append(rows, StyleRowSelected.Width(m.width-2).Render(rowContent))
		} else {
			rows = append(rows, StyleRow.Width(m.width-2).Render(rowContent))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderDeviceDetails() string {
	if len(m.devices) == 0 {
		return lipgloss.NewStyle().Foreground(ColorMuted).Padding(2, 4).Render("No wireless devices detected.")
	}

	dev := m.devices[m.selectedDev]

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Margin(1, 0)

	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorHighlight).Width(20)
	valStyle := lipgloss.NewStyle().Foreground(ColorText)

	renderRow := func(label, val string) string {
		return lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render(label), valStyle.Render(val))
	}

	powerStatus := "Enabled"
	if !dev.Powered {
		powerStatus = "Disabled (Powered Off)"
	}

	stationState := "Disconnected"
	if m.station != nil && m.station.State != "" {
		stationState = m.station.State
	}

	connectedName := "None"
	if m.activeNetwork != nil {
		connectedName = fmt.Sprintf("%s (%4.1f dBm)", m.activeNetwork.Name, m.activeNetwork.SignalDBm())
	}

	fields := []string{
		lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("📡 Wireless Device Information"),
		"",
		renderRow("Interface Name:", dev.Name),
		renderRow("MAC Address:", dev.Address),
		renderRow("Power State:", powerStatus),
		renderRow("Operating Mode:", dev.Mode),
		renderRow("Station State:", stationState),
		renderRow("Current Network:", connectedName),
		renderRow("Adapter PHY:", dev.AdapterName),
		renderRow("Chipset Model:", dev.AdapterModel),
		renderRow("Vendor:", dev.AdapterVendor),
	}

	return cardStyle.Render(lipgloss.JoinVertical(lipgloss.Left, fields...))
}

func (m Model) renderFooter() string {
	// 1. Status / Toast message line
	var statusLine string
	if m.busyAction != "" {
		statusLine = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).Render(
			fmt.Sprintf("%s %s", m.spinner.View(), m.busyAction),
		)
	} else if m.notification != nil && time.Now().Before(m.notification.Expires) {
		var style lipgloss.Style
		switch m.notification.Type {
		case NotifSuccess:
			style = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
		case NotifError:
			style = lipgloss.NewStyle().Foreground(ColorDanger).Bold(true)
		case NotifWarning:
			style = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
		default:
			style = lipgloss.NewStyle().Foreground(ColorHighlight)
		}
		statusLine = style.Render(m.notification.Message)
	}

	// 2. Keybinding helper row
	k := func(key, desc string) string {
		return lipgloss.JoinHorizontal(lipgloss.Top,
			StyleHelpKey.Render(key),
			" ",
			StyleHelpDesc.Render(desc),
		)
	}

	var shortcuts []string
	if m.activeTab == TabAvailable {
		shortcuts = []string{
			k("[Enter]", "connect"),
			k("[d]", "disconnect"),
			k("[s]", "scan"),
			k("[f]", "forget"),
			k("[p]", "power"),
			k("[/]", "filter"),
			k("[Tab]", "next tab"),
			k("[?]", "help"),
			k("[q]", "quit"),
		}
	} else if m.activeTab == TabSaved {
		shortcuts = []string{
			k("[Enter]", "connect"),
			k("[f]", "forget"),
			k("[a]", "autoconnect"),
			k("[/]", "filter"),
			k("[Tab]", "next tab"),
			k("[?]", "help"),
			k("[q]", "quit"),
		}
	} else {
		shortcuts = []string{
			k("[p]", "toggle power"),
			k("[Tab]", "next tab"),
			k("[?]", "help"),
			k("[q]", "quit"),
		}
	}

	helpRow := strings.Join(shortcuts, "  ")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		StyleStatusBar.Width(m.width-2).Render(statusLine),
		lipgloss.NewStyle().Padding(0, 1).Render(helpRow),
	)
}
