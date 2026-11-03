package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var pageStyle = lipgloss.NewStyle().Padding(1, 2)

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing nm-go..."
	}

	// Active modal overlays
	if m.modal != ModalNone {
		switch m.modal {
		case ModalPassword:
			return pageStyle.Render(renderPasswordModal(m.passwordModal, m.currentTheme, m.width-4))
		case ModalConfirmForget:
			name := "selected network"
			if m.targetForget != nil {
				name = m.targetForget.Name
			}
			return pageStyle.Render(renderConfirmForgetModal(name, m.currentTheme, m.width-4))
		case ModalHelp:
			return pageStyle.Render(renderHelpModal(m.currentTheme, m.width-4))
		case ModalThemePicker:
			return pageStyle.Render(renderThemePickerModal(m.currentTheme, m.themePickerIndex, m.width-4))
		}
	}

	totalWidth := m.width - 4
	if totalWidth < 50 {
		totalWidth = 50
	}

	// 1. Header Banner Pill & Tabs
	headerBanner := m.renderHeaderBanner()
	tabsRow := m.renderTabs()

	// 2. Main Content Container
	contentView := m.renderMainTable(totalWidth)

	// 3. Optional Search Box (UniSocket-style)
	var searchBox string
	if m.filtering || m.filterInput.Value() != "" {
		searchBox = "\n" + m.renderSearchBox(totalWidth)
	}

	// 4. Bottom Deck (Commands Box + Monitor Box)
	bottomDeck := "\n" + m.renderBottomDeck(totalWidth)

	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		headerBanner,
		tabsRow,
		contentView,
		searchBox,
		bottomDeck,
	)

	return pageStyle.Render(fullContent)
}

func (m Model) renderHeaderBanner() string {
	titlePill := m.currentTheme.TitleStyle().Render("nm-go  •  Wireless Network Manager")

	var stateText string
	if m.isScanning {
		stateText = m.spinner.View() + " Scanning airwaves..."
	} else if m.isConnecting {
		stateText = m.spinner.View() + " " + m.busyAction
	} else if m.activeNetwork != nil {
		stateText = fmt.Sprintf("Connected: %s (%d%%)", m.activeNetwork.Name, m.activeNetwork.SignalPercent())
	} else {
		stateText = "Disconnected"
	}

	stateStyled := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.currentTheme.Status).
		Render(stateText)

	return lipgloss.JoinHorizontal(lipgloss.Center, titlePill, "   ", stateStyled) + "\n"
}

func (m Model) renderTabs() string {
	availCount := len(m.networks)
	savedCount := len(m.knownNetworks)

	tab0 := RenderTab(m.currentTheme, fmt.Sprintf("Available (%d)", availCount), m.activeTab == TabAvailable)
	tab1 := RenderTab(m.currentTheme, fmt.Sprintf("Saved (%d)", savedCount), m.activeTab == TabSaved)
	tab2 := RenderTab(m.currentTheme, "Device Diagnostics", m.activeTab == TabDetails)

	return lipgloss.JoinHorizontal(lipgloss.Top, tab0, tab1, tab2) + "\n"
}

func (m Model) renderMainTable(totalWidth int) string {
	tableContainer := m.currentTheme.TableContainerStyle().Width(totalWidth - 2)

	switch m.activeTab {
	case TabAvailable:
		return tableContainer.Render(m.renderAvailableTab(totalWidth - 4))
	case TabSaved:
		return tableContainer.Render(m.renderSavedTab(totalWidth - 4))
	case TabDetails:
		return tableContainer.Render(m.renderDetailsTab(totalWidth - 4))
	default:
		return ""
	}
}

func (m Model) renderAvailableTab(innerW int) string {
	networks := m.getFilteredNetworks()

	colWName := 30
	colWSec := 20
	colWSig := 26
	colWStatus := 14

	if innerW > 95 {
		colWName = 36
		colWSec = 22
		colWSig = 28
		colWStatus = innerW - colWName - colWSec - colWSig - 6
	}

	headerText := fmt.Sprintf("  %-*s %-*s %-*s %-*s",
		colWName, "SSID",
		colWSec, "SECURITY",
		colWSig, "SIGNAL",
		colWStatus, "STATUS",
	)
	header := m.currentTheme.HeaderStyle().Width(innerW).Render(headerText)

	if len(networks) == 0 {
		emptyMsg := "No networks in range. Press [s] to trigger a scan."
		if m.filterInput.Value() != "" {
			emptyMsg = fmt.Sprintf("No networks matching '%s'. Press [Esc] to clear.", m.filterInput.Value())
		}
		emptyBlock := lipgloss.NewStyle().
			Foreground(m.currentTheme.Muted).
			Padding(2, 2).
			Render(emptyMsg)
		return lipgloss.JoinVertical(lipgloss.Left, header, emptyBlock)
	}

	maxVisible := m.height - 15
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
	rows = append(rows, header)

	for i := start; i < end; i++ {
		net := networks[i]
		isSelected := i == m.cursorAvailable

		cursor := "  "
		if isSelected {
			cursor = "▸ "
		}

		nameStr := net.Name
		if len(nameStr) > colWName-2 {
			nameStr = nameStr[:colWName-5] + "..."
		}

		secStr := net.SecurityLabel()
		if len(secStr) > colWSec-2 {
			secStr = secStr[:colWSec-5] + "..."
		}

		pct := net.SignalPercent()
		sigMeter := RenderSignalMeter(m.currentTheme, pct, net.SignalDBm())

		statusText := ""
		if net.Connected {
			statusText = lipgloss.NewStyle().Bold(true).Foreground(m.currentTheme.Success).Render("✔ Connected")
		} else if net.Known {
			statusText = lipgloss.NewStyle().Foreground(m.currentTheme.Highlight).Render("★ Saved")
		} else if net.Type == "open" {
			statusText = lipgloss.NewStyle().Foreground(m.currentTheme.Warning).Render("● Open")
		}

		rowContent := fmt.Sprintf("%s%-*s %-*s %s %s",
			cursor,
			colWName-2, nameStr,
			colWSec, secStr,
			sigMeter,
			statusText,
		)

		if isSelected {
			rows = append(rows, m.currentTheme.SelectedRowStyle().Width(innerW).Render(rowContent))
		} else {
			rows = append(rows, m.currentTheme.RowStyle().Width(innerW).Render(rowContent))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderSavedTab(innerW int) string {
	known := m.getFilteredKnownNetworks()

	colWName := 32
	colWSec := 20
	colWAuto := 16
	colWLast := 20

	if innerW > 95 {
		colWName = 38
		colWSec = 22
		colWAuto = 18
		colWLast = innerW - colWName - colWSec - colWAuto - 6
	}

	headerText := fmt.Sprintf("  %-*s %-*s %-*s %-*s",
		colWName, "SSID",
		colWSec, "SECURITY",
		colWAuto, "AUTO-CONNECT",
		colWLast, "LAST CONNECTED",
	)
	header := m.currentTheme.HeaderStyle().Width(innerW).Render(headerText)

	if len(known) == 0 {
		emptyMsg := "No saved networks found."
		if m.filterInput.Value() != "" {
			emptyMsg = fmt.Sprintf("No saved networks matching '%s'.", m.filterInput.Value())
		}
		emptyBlock := lipgloss.NewStyle().
			Foreground(m.currentTheme.Muted).
			Padding(2, 2).
			Render(emptyMsg)
		return lipgloss.JoinVertical(lipgloss.Left, header, emptyBlock)
	}

	maxVisible := m.height - 15
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
	rows = append(rows, header)

	for i := start; i < end; i++ {
		kn := known[i]
		isSelected := i == m.cursorSaved

		cursor := "  "
		if isSelected {
			cursor = "▸ "
		}

		nameStr := kn.Name
		if len(nameStr) > colWName-2 {
			nameStr = nameStr[:colWName-5] + "..."
		}

		secStr := kn.Type
		if secStr == "psk" {
			secStr = "WPA2/WPA3 (PSK)"
		}

		autoStr := "No"
		if kn.AutoConnect {
			autoStr = lipgloss.NewStyle().Foreground(m.currentTheme.Success).Bold(true).Render("✔ Yes")
		} else {
			autoStr = lipgloss.NewStyle().Foreground(m.currentTheme.Muted).Render("✘ No")
		}

		lastStr := kn.FormattedLastConnected()

		rowContent := fmt.Sprintf("%s%-*s %-*s %-*s %s",
			cursor,
			colWName-2, nameStr,
			colWSec, secStr,
			colWAuto+4, autoStr,
			lastStr,
		)

		if isSelected {
			rows = append(rows, m.currentTheme.SelectedRowStyle().Width(innerW).Render(rowContent))
		} else {
			rows = append(rows, m.currentTheme.RowStyle().Width(innerW).Render(rowContent))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderDetailsTab(innerW int) string {
	if len(m.devices) == 0 {
		return lipgloss.NewStyle().Foreground(m.currentTheme.Muted).Padding(2, 2).Render("No wireless hardware detected.")
	}

	dev := m.devices[m.selectedDev]

	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(m.currentTheme.Highlight).Width(22)
	valStyle := lipgloss.NewStyle().Foreground(m.currentTheme.Text)

	row := func(label, val string) string {
		return lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render(label), valStyle.Render(val))
	}

	powerStatus := "Enabled"
	if !dev.Powered {
		powerStatus = "Disabled (Radio Off)"
	}

	stationState := "Disconnected"
	if m.station != nil && m.station.State != "" {
		stationState = m.station.State
	}

	connectedName := "None"
	if m.activeNetwork != nil {
		connectedName = fmt.Sprintf("%s (%4.1f dBm)", m.activeNetwork.Name, m.activeNetwork.SignalDBm())
	}

	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(m.currentTheme.Highlight).Render("📡 Hardware & Link Diagnostics"),
		"",
		row("Interface Name:", dev.Name),
		row("MAC Address:", dev.Address),
		row("Radio Power:", powerStatus),
		row("Interface Mode:", dev.Mode),
		row("Station Status:", stationState),
		row("Connected Network:", connectedName),
		row("Physical Adapter:", dev.AdapterName),
		row("Chipset Hardware:", dev.AdapterModel),
		row("Hardware Vendor:", dev.AdapterVendor),
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) renderSearchBox(totalWidth int) string {
	card := m.currentTheme.CardStyle().Width(totalWidth - 2)

	title := lipgloss.NewStyle().Bold(true).Foreground(m.currentTheme.Highlight).Render("Search Networks")
	inputLine := "/ " + m.filterInput.View()
	hints := lipgloss.NewStyle().Foreground(m.currentTheme.Muted).Render(
		"Esc / Ctrl+C: Cancel & Clear  •  Enter: Keep Filter  •  Tab / ↓: Move to Table",
	)

	return card.Render(lipgloss.JoinVertical(lipgloss.Left, title, inputLine, hints))
}

func (m Model) renderBottomDeck(totalWidth int) string {
	if totalWidth >= 78 {
		rightWidth := 34
		leftWidth := totalWidth - rightWidth - 2
		leftCard := m.renderCommandsCard(leftWidth)
		rightCard := m.renderMonitorCard(rightWidth)
		return lipgloss.JoinHorizontal(lipgloss.Top, leftCard, "  ", rightCard)
	}

	// Narrow fallback
	return m.renderCommandsCard(totalWidth)
}

func (m Model) renderCommandsCard(width int) string {
	card := m.currentTheme.CardStyle().Width(width - 2).Height(4)

	title := lipgloss.NewStyle().Bold(true).Foreground(m.currentTheme.Highlight).Render("Commands")

	b := func(k, a string) string {
		return RenderCommandBadge(m.currentTheme, k, a)
	}

	row1 := fmt.Sprintf("%s   %s   %s   %s   %s",
		b("↑/↓", "Nav"),
		b("Enter", "Connect"),
		b("d", "Disconnect"),
		b("s", "Scan"),
		b("f", "Forget"),
	)

	row2 := fmt.Sprintf("%s   %s   %s   %s   %s",
		b("p", "Power"),
		b("/", "Search"),
		b("t", "Theme"),
		b("?", "Help"),
		b("q", "Quit"),
	)

	tip := lipgloss.NewStyle().Foreground(m.currentTheme.Muted).Render(
		"Tip: Press [t] to switch theme • [Tab] to switch view • [a] toggle autoconnect",
	)

	return card.Render(lipgloss.JoinVertical(lipgloss.Left, title, row1, row2, tip))
}

func (m Model) renderMonitorCard(width int) string {
	card := m.currentTheme.CardStyle().Width(width - 2).Height(4)

	title := lipgloss.NewStyle().Bold(true).Foreground(m.currentTheme.Highlight).Render("Monitor")

	labelStyle := lipgloss.NewStyle().Foreground(m.currentTheme.Muted)
	valStyle := lipgloss.NewStyle().Bold(true).Foreground(m.currentTheme.Text)

	themeName := m.currentTheme.Name
	line1 := labelStyle.Render("Theme: ") + valStyle.Render(themeName)

	devName := "wlan0"
	if len(m.devices) > 0 {
		devName = m.devices[0].Name
	}
	stateStr := "Disconnected"
	if m.activeNetwork != nil {
		stateStr = m.activeNetwork.Name
	}
	line2 := labelStyle.Render("Device: ") + valStyle.Render(devName) +
		labelStyle.Render(" • ") +
		lipgloss.NewStyle().Foreground(m.currentTheme.Status).Render(stateStr)

	var line3 string
	if m.notification != nil && time.Now().Before(m.notification.Expires) {
		line3 = lipgloss.NewStyle().Foreground(m.currentTheme.Highlight).Bold(true).Render("● " + m.notification.Message)
	} else if m.busyAction != "" {
		line3 = lipgloss.NewStyle().Foreground(m.currentTheme.Warning).Bold(true).Render("● " + m.busyAction)
	} else {
		line3 = lipgloss.NewStyle().Foreground(m.currentTheme.Success).Render("● Live feed (3s)")
	}

	return card.Render(lipgloss.JoinVertical(lipgloss.Left, title, line1, line2, line3))
}
