package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/Senasphy/nm-go/internal/theme"
)

var pageStyle = lipgloss.NewStyle().Padding(1, 2)

const (
	minMainHeight = 15
	themeBoxWidth = 26
)

func (m Model) getMainSectionHeight() int {
	h := minMainHeight
	if m.height > 24 {
		h = m.height - 12
	}
	if h < minMainHeight {
		h = minMainHeight
	}
	return h
}

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing nm-go..."
	}

	// Active modal overlays (only for forms like Password / Confirm / Help)
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
		}
	}

	totalWidth := m.width - 4
	if totalWidth < 50 {
		totalWidth = 50
	}

	mainH := m.getMainSectionHeight()

	// 1. Header Banner Pill & Live Connection Status
	headerBanner := m.renderHeaderBanner()
	tabsRow := m.renderTabs()

	// 2. Main Section: Table (and optional side-by-side Theme Switcher) with fixed, balanced height
	var mainSection string
	if m.isThemePicker && totalWidth >= 75 {
		tableW := totalWidth - themeBoxWidth - 2
		tableView := m.renderMainTable(tableW, mainH)
		themeBox := m.renderSideThemePicker(themeBoxWidth, mainH)
		mainSection = lipgloss.JoinHorizontal(lipgloss.Top, tableView, "  ", themeBox)
	} else {
		mainSection = m.renderMainTable(totalWidth, mainH)
	}

	// 3. Optional Search Box
	var searchBox string
	if m.filtering || m.filterInput.Value() != "" {
		searchBox = "\n" + m.renderSearchBox(totalWidth)
	}

	// 4. Bottom Commands Deck (Clean, single full-width card - no cluttered monitor box)
	bottomDeck := "\n" + m.renderCommandsCard(totalWidth)

	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		headerBanner,
		tabsRow,
		mainSection,
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
		pct := m.activeNetwork.SignalPercent()
		stateText = fmt.Sprintf("Connected: %s (%d%%)", m.activeNetwork.Name, pct)
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

func (m Model) renderMainTable(totalWidth int, height int) string {
	tableContainer := m.currentTheme.TableContainerStyle().
		Width(totalWidth - 2).
		Height(height)

	switch m.activeTab {
	case TabAvailable:
		return tableContainer.Render(m.renderAvailableTab(totalWidth-4, height))
	case TabSaved:
		return tableContainer.Render(m.renderSavedTab(totalWidth-4, height))
	case TabDetails:
		return tableContainer.Render(m.renderDetailsTab(totalWidth-4, height))
	default:
		return ""
	}
}

func (m Model) renderAvailableTab(innerW int, height int) string {
	networks := m.getFilteredNetworks()

	colWName := 28
	colWSec := 18
	colWSig := 24
	colWStatus := 14

	if innerW > 85 {
		colWName = 34
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

	maxVisible := height - 1
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

func (m Model) renderSavedTab(innerW int, height int) string {
	known := m.getFilteredKnownNetworks()

	colWName := 30
	colWSec := 18
	colWAuto := 16
	colWLast := 18

	if innerW > 85 {
		colWName = 36
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

	maxVisible := height - 1
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

func (m Model) renderDetailsTab(innerW int, height int) string {
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

// renderSideThemePicker renders a narrow, clean side box matching the main table height
func (m Model) renderSideThemePicker(width int, height int) string {
	card := m.currentTheme.CardStyle().
		Width(width - 2).
		Height(height)

	title := lipgloss.NewStyle().Bold(true).Foreground(m.currentTheme.Highlight).Render("Theme")

	themes := theme.List()
	cursorStyle := lipgloss.NewStyle().Bold(true).Foreground(m.currentTheme.Highlight)
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(m.currentTheme.Success)
	itemStyle := lipgloss.NewStyle().Foreground(m.currentTheme.Text)

	items := make([]string, len(themes))
	for i, t := range themes {
		cursor := "  "
		if i == m.themePickerIndex {
			cursor = "▸ "
		}

		indicator := "  "
		if t.ID == m.currentTheme.ID {
			indicator = "✓ "
		}

		nameStr := t.Name
		if len(nameStr) > width-8 {
			nameStr = nameStr[:width-10] + ".."
		}

		if i == m.themePickerIndex {
			items[i] = cursorStyle.Render(cursor) + selectedStyle.Render(indicator) + cursorStyle.Render(nameStr)
		} else {
			items[i] = itemStyle.Render(cursor) + selectedStyle.Render(indicator) + itemStyle.Render(nameStr)
		}
	}

	hint := lipgloss.NewStyle().Foreground(m.currentTheme.Muted).Render("[Enter/t] Close")

	content := append([]string{title, ""}, items...)
	content = append(content, "", hint)

	return card.Render(lipgloss.JoinVertical(lipgloss.Left, content...))
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

// renderCommandsCard renders the full-width commands card without the cluttered monitor box
func (m Model) renderCommandsCard(width int) string {
	card := m.currentTheme.CardStyle().Width(width - 2)

	title := lipgloss.NewStyle().Bold(true).Foreground(m.currentTheme.Highlight).Render("Commands")

	b := func(k, a string) string {
		return RenderCommandBadge(m.currentTheme, k, a)
	}

	row1 := fmt.Sprintf("%s   %s   %s   %s   %s   %s   %s   %s   %s   %s",
		b("↑/↓", "Nav"),
		b("Enter", "Connect"),
		b("d", "Disconnect"),
		b("s", "Scan"),
		b("f", "Forget"),
		b("p", "Power"),
		b("/", "Search"),
		b("t", "Theme"),
		b("?", "Help"),
		b("q", "Quit"),
	)

	var statusLine string
	if m.notification != nil && time.Now().Before(m.notification.Expires) {
		statusLine = lipgloss.NewStyle().Foreground(m.currentTheme.Status).Bold(true).Render("● " + m.notification.Message)
	} else if m.busyAction != "" {
		statusLine = lipgloss.NewStyle().Foreground(m.currentTheme.Warning).Bold(true).Render("● " + m.busyAction)
	} else {
		statusLine = lipgloss.NewStyle().Foreground(m.currentTheme.Muted).Render(
			"Tip: Press [t] to toggle theme panel • [Tab] to switch view • [a] autoconnect",
		)
	}

	return card.Render(lipgloss.JoinVertical(lipgloss.Left, title, row1, statusLine))
}
