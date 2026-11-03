package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"nm-go/internal/iwd"
)

type ModalType int

const (
	ModalNone ModalType = iota
	ModalPassword
	ModalConfirmForget
	ModalHelp
)

type PasswordModal struct {
	Network   iwd.Network
	Input     textinput.Model
	ShowClear bool
}

func NewPasswordModal(net iwd.Network) PasswordModal {
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

func renderPasswordModal(pm PasswordModal, width int) string {
	title := StyleModalTitle.Render(fmt.Sprintf("🔑 Connect to %s", pm.Network.Name))

	securityInfo := StyleBadgeMuted.Render(pm.Network.SecurityLabel())

	eyeIcon := "[Ctrl+R] Show"
	if pm.ShowClear {
		eyeIcon = "[Ctrl+R] Hide"
	}

	hint := lipgloss.NewStyle().Foreground(ColorMuted).Render(eyeIcon)
	inputBox := lipgloss.JoinHorizontal(lipgloss.Center, pm.Input.View(), " ", hint)

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render("[Enter] Connect"),
		"    ",
		lipgloss.NewStyle().Foreground(ColorDanger).Render("[Esc] Cancel"),
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

	boxWidth := 48
	if width > 0 && width < 52 {
		boxWidth = width - 4
	}

	modalBox := StyleModalBorder.Width(boxWidth).Render(body)
	return lipgloss.Place(width, 10, lipgloss.Center, lipgloss.Center, modalBox)
}

func renderConfirmForgetModal(name string, width int) string {
	title := StyleModalTitle.Render("⚠️  Forget Network")
	msg := fmt.Sprintf("Are you sure you want to forget '%s'?", name)
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render("[y] Yes, Forget"),
		"    ",
		lipgloss.NewStyle().Foreground(ColorMuted).Render("[n/Esc] Cancel"),
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

	modalBox := StyleModalBorder.Width(boxWidth).Render(body)
	return lipgloss.Place(width, 10, lipgloss.Center, lipgloss.Center, modalBox)
}

func renderHelpModal(width int) string {
	title := StyleModalTitle.Render("📖 Keyboard Shortcuts & Help")

	keyPairs := [][]string{
		{"Tab / Shift+Tab", "Switch tabs (Available / Saved / Info)"},
		{"↑ / k , ↓ / j", "Navigate network list"},
		{"Enter / c", "Connect to selected network"},
		{"d", "Disconnect current network"},
		{"s / r", "Scan & refresh Wi-Fi networks"},
		{"f / x", "Forget saved network credentials"},
		{"a", "Toggle auto-connect for saved network"},
		{"p", "Toggle Wi-Fi power (radio on/off)"},
		{"/", "Filter networks by SSID"},
		{"Esc", "Cancel / close dialog / clear filter"},
		{"?", "Toggle this help screen"},
		{"q / Ctrl+C", "Quit nm-go"},
	}

	var rows []string
	for _, pair := range keyPairs {
		k := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Width(18).Render(pair[0])
		v := lipgloss.NewStyle().Foreground(ColorText).Render(pair[1])
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, k, v))
	}

	hint := lipgloss.NewStyle().Foreground(ColorMuted).MarginTop(1).Render("Press [Esc] or [?] to close")
	body := lipgloss.JoinVertical(lipgloss.Left, append([]string{title, ""}, append(rows, hint)...)...)

	boxWidth := 56
	if width > 0 && width < 60 {
		boxWidth = width - 4
	}

	modalBox := StyleModalBorder.Width(boxWidth).Render(body)
	return lipgloss.Place(width, 16, lipgloss.Center, lipgloss.Center, modalBox)
}
