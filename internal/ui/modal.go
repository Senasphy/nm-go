package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"nm-go/internal/iwd"
	"nm-go/internal/theme"
)

type ModalType int

const (
	ModalNone ModalType = iota
	ModalPassword
	ModalConfirmForget
	ModalHelp
	ModalThemePicker
)

type PasswordModal struct {
	Network   iwd.Network
	Input     textinput.Model
	ShowClear bool
}

func NewPasswordModal(net iwd.Network, th theme.Theme) PasswordModal {
	ti := textinput.New()
	ti.Placeholder = "Enter passphrase..."
	ti.Focus()
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.CharLimit = 64
	ti.Width = 36

	return PasswordModal{
		Network:   net,
		Input:     ti,
		ShowClear: false,
	}
}

func (pm *PasswordModal) ToggleEcho() {
	pm.ShowClear = !pm.ShowClear
	if pm.ShowClear {
		pm.Input.EchoMode = textinput.EchoNormal
	} else {
		pm.Input.EchoMode = textinput.EchoPassword
	}
}

func renderPasswordModal(pm PasswordModal, th theme.Theme, width int) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(th.Highlight).Render(fmt.Sprintf("🔑 Connect to %s", pm.Network.Name))

	securityInfo := lipgloss.NewStyle().
		Bold(true).
		Foreground(th.TitleText).
		Background(th.Muted).
		Padding(0, 1).
		Render(pm.Network.SecurityLabel())

	eyeIcon := "[Ctrl+R] Show"
	if pm.ShowClear {
		eyeIcon = "[Ctrl+R] Hide"
	}

	hint := lipgloss.NewStyle().Foreground(th.Muted).Render(eyeIcon)
	inputBox := lipgloss.JoinHorizontal(lipgloss.Center, pm.Input.View(), " ", hint)

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Bold(true).Foreground(th.Success).Render("[Enter] Connect"),
		"    ",
		lipgloss.NewStyle().Foreground(th.Danger).Render("[Esc] Cancel"),
	)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		securityInfo,
		"",
		"Passphrase:",
		inputBox,
		"",
		buttons,
	)

	boxWidth := 50
	if width > 0 && width < 54 {
		boxWidth = width - 4
	}

	modalBox := th.ModalStyle().Width(boxWidth).Render(body)
	return lipgloss.Place(width, 11, lipgloss.Center, lipgloss.Center, modalBox)
}

func renderConfirmForgetModal(name string, th theme.Theme, width int) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(th.Danger).Render("⚠️  Forget Network")
	msg := fmt.Sprintf("Are you sure you want to forget '%s'?", name)
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Bold(true).Foreground(th.Danger).Render("[y] Yes, Forget"),
		"    ",
		lipgloss.NewStyle().Foreground(th.Muted).Render("[n/Esc] Cancel"),
	)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		msg,
		"",
		buttons,
	)

	boxWidth := 48
	if width > 0 && width < 52 {
		boxWidth = width - 4
	}

	modalBox := th.ModalStyle().Width(boxWidth).Render(body)
	return lipgloss.Place(width, 9, lipgloss.Center, lipgloss.Center, modalBox)
}

func renderHelpModal(th theme.Theme, width int) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(th.Highlight).Render("📖 Keyboard Shortcuts & Commands")

	keyPairs := [][]string{
		{"Tab / Shift+Tab", "Switch tabs (Available / Saved / Info)"},
		{"↑ / k , ↓ / j", "Navigate network list"},
		{"Enter / c", "Connect to selected network"},
		{"d", "Disconnect active Wi-Fi network"},
		{"s / r", "Scan & refresh Wi-Fi networks"},
		{"f / x", "Forget saved network credentials"},
		{"a", "Toggle auto-connect for saved network"},
		{"p", "Toggle Wi-Fi radio power on/off"},
		{"/", "Search & filter networks by SSID"},
		{"t", "Open interactive theme switcher"},
		{"Esc", "Cancel / close dialog / clear filter"},
		{"?", "Toggle this help screen"},
		{"q / Ctrl+C", "Quit nm-go"},
	}

	var rows []string
	for _, pair := range keyPairs {
		k := lipgloss.NewStyle().Bold(true).Foreground(th.Highlight).Width(18).Render(pair[0])
		v := lipgloss.NewStyle().Foreground(th.Text).Render(pair[1])
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, k, v))
	}

	hint := lipgloss.NewStyle().Foreground(th.Muted).MarginTop(1).Render("Press [Esc] or [?] to close")
	body := lipgloss.JoinVertical(lipgloss.Left, append([]string{title, ""}, append(rows, hint)...)...)

	boxWidth := 58
	if width > 0 && width < 62 {
		boxWidth = width - 4
	}

	modalBox := th.ModalStyle().Width(boxWidth).Render(body)
	return lipgloss.Place(width, 17, lipgloss.Center, lipgloss.Center, modalBox)
}

func renderThemePickerModal(th theme.Theme, selectedIdx int, width int) string {
	themes := theme.List()
	title := lipgloss.NewStyle().Bold(true).Foreground(th.Highlight).Render("🎨 Choose Theme")

	cursorStyle := lipgloss.NewStyle().Bold(true).Foreground(th.Highlight)
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(th.Success)
	itemStyle := lipgloss.NewStyle().Foreground(th.Text)
	descStyle := lipgloss.NewStyle().Foreground(th.Muted)

	items := make([]string, len(themes))
	for i, t := range themes {
		cursor := "  "
		if i == selectedIdx {
			cursor = "▸ "
		}

		indicator := "  "
		if t.ID == th.ID {
			indicator = "✓ "
		}

		nameCol := fmt.Sprintf("%-18s", t.Name)
		if i == selectedIdx {
			items[i] = cursorStyle.Render(cursor) + selectedStyle.Render(indicator) + cursorStyle.Render(nameCol) + " " + descStyle.Render(t.Description)
		} else {
			items[i] = itemStyle.Render(cursor) + selectedStyle.Render(indicator) + itemStyle.Render(nameCol) + " " + descStyle.Render(t.Description)
		}
	}

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Bold(true).Foreground(th.Success).Render("[Enter] Apply"),
		"    ",
		lipgloss.NewStyle().Foreground(th.Muted).Render("[Esc] Cancel"),
	)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		lipgloss.JoinVertical(lipgloss.Left, items...),
		"",
		buttons,
	)

	boxWidth := 60
	if width > 0 && width < 64 {
		boxWidth = width - 4
	}

	modalBox := th.ModalStyle().Width(boxWidth).Render(body)
	return lipgloss.Place(width, len(themes)+6, lipgloss.Center, lipgloss.Center, modalBox)
}
