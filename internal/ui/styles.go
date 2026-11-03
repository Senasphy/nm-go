package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"nm-go/internal/theme"
)

// SignalMeter renders a btop-inspired visual bar meter with percentage and dBm
func RenderSignalMeter(th theme.Theme, pct int, dbm float64) string {
	bars := 8
	filled := (pct * bars) / 100
	if filled > bars {
		filled = bars
	}
	if filled < 1 && pct > 0 {
		filled = 1
	}

	meterStr := strings.Repeat("■", filled) + strings.Repeat("·", bars-filled)

	var color lipgloss.TerminalColor
	switch {
	case pct >= 65:
		color = th.Success
	case pct >= 35:
		color = th.Warning
	default:
		color = th.Danger
	}

	meterStyled := lipgloss.NewStyle().Foreground(color).Bold(true).Render(meterStr)
	textStyled := lipgloss.NewStyle().Foreground(th.Text).Render(fmt.Sprintf("%3d%% (%4.1fdBm)", pct, dbm))

	return meterStyled + " " + textStyled
}

// RenderCommandBadge renders an inverted keycap pill like in UniSocket
func RenderCommandBadge(th theme.Theme, key, action string) string {
	kStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(th.TitleText).
		Background(th.Highlight).
		Padding(0, 1).
		Render(key)

	aStyle := lipgloss.NewStyle().
		Foreground(th.Text).
		Render(" " + action)

	return kStyle + aStyle
}

// RenderTab renders a styled tab pill
func RenderTab(th theme.Theme, title string, active bool) string {
	if active {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(th.TitleText).
			Background(th.Highlight).
			Padding(0, 2).
			MarginRight(1).
			Render(title)
	}

	return lipgloss.NewStyle().
		Foreground(th.Muted).
		Background(th.HeaderBg).
		Padding(0, 2).
		MarginRight(1).
		Render(title)
}
