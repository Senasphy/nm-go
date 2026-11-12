package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/Senasphy/nm-go/internal/iwd"
	"github.com/Senasphy/nm-go/internal/theme"
)

type ModalType int

const (
	ModalNone ModalType = iota
	ModalPassword
	ModalConfirmForget
	ModalHelp
)

type PasswordModal struct {
	Network     iwd.Network
	Input       textinput.Model
	ShowPass    bool
	Err         string
	borderColor lipgloss.TerminalColor
}

func NewPasswordModal(net iwd.Network, th theme.Theme) PasswordModal {
	ti := textinput.New()
	ti.Placeholder = "Enter WPA passphrase..."
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 36

	return PasswordModal{
		Network:     net,
		Input:       ti,
		ShowPass:    false,
		borderColor: th.Highlight,
	}
}

func (p *PasswordModal) ToggleEcho() {
	p.ShowPass = !p.ShowPass
	if p.ShowPass {
		p.Input.EchoMode = textinput.EchoNormal
	} else {
		p.Input.EchoMode = textinput.EchoPassword
	}
}

func renderPasswordModal(p PasswordModal, th theme.Theme, width int) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(th.Highlight).Render("🔐  Connect to " + p.Network.Name)
	securityInfo := lipgloss.NewStyle().Foreground(th.Muted).Render("Security: " + p.Network.SecurityLabel())

	inputBox := p.Input.View()

	echoHint := "[Ctrl+R] Show/Hide password"
	if p.ShowPass {
		echoHint = "[Ctrl+R] Hide password"
	}
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Foreground(th.Highlight).Render("[Enter] Connect"),
		"    ",
		lipgloss.NewStyle().Foreground(th.Muted).Render(echoHint),
		"    ",
		lipgloss.NewStyle().Foreground(th.Muted).Render("[Esc] Cancel"),
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
		{"Tab / Shift+Tab", "Switch tabs (Available / Saved / Ethernet / Info)"},
		{"↑ / k , ↓ / j", "Navigate network / interface list / themes"},
		{"Enter / c", "Connect to network / activate interface"},
		{"d", "Disconnect active network / interface"},
		{"s / r", "Scan Wi-Fi & refresh network interfaces"},
		{"f / x", "Forget saved Wi-Fi network credentials"},
		{"a", "Toggle auto-connect for saved network"},
		{"p", "Toggle Wi-Fi radio power on/off"},
		{"/", "Search & filter networks and interfaces"},
		{"t", "Toggle side-by-side theme switcher"},
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
	return lipgloss.Place(width, 18, lipgloss.Center, lipgloss.Center, modalBox)
}
