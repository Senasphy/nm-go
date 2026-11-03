package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Theme palette
	ColorPrimary   = lipgloss.Color("#00D4FF")
	ColorSecondary = lipgloss.Color("#A855F7")
	ColorSuccess   = lipgloss.Color("#22C55E")
	ColorWarning   = lipgloss.Color("#EAB308")
	ColorDanger    = lipgloss.Color("#EF4444")
	ColorMuted     = lipgloss.Color("#6B7280")
	ColorHighlight = lipgloss.Color("#38BDF8")
	ColorDarkBg    = lipgloss.Color("#1E1E2E")
	ColorCardBg    = lipgloss.Color("#262637")
	ColorText      = lipgloss.Color("#F3F4F6")

	// App Header
	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#4338CA")).
			Padding(0, 1)

	StyleBadge = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			MarginRight(1)

	StyleBadgeSuccess = StyleBadge.
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(ColorSuccess)

	StyleBadgeWarning = StyleBadge.
				Foreground(lipgloss.Color("#000000")).
				Background(ColorWarning)

	StyleBadgeDanger = StyleBadge.
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(ColorDanger)

	StyleBadgeMuted = StyleBadge.
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorMuted)

	// Tabs
	StyleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1).
			MarginRight(1)

	StyleTabInactive = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorMuted).
				Padding(0, 1).
				MarginRight(1)

	// Tables and Rows
	StyleTableHeader = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorHighlight).
				Padding(0, 1)

	StyleRow = lipgloss.NewStyle().
			Padding(0, 1)

	StyleRowSelected = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorText).
				Background(lipgloss.Color("#374151")).
				Padding(0, 1)

	StyleConnectedTag = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorSuccess)

	StyleKnownTag = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	// Signal meters
	StyleSignalHigh = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	StyleSignalMid = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	StyleSignalLow = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	// Status and Notification Bar
	StyleStatusBar = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("#374151")).
			Padding(0, 1).
			MarginTop(1)

	StyleNotification = lipgloss.NewStyle().
				Padding(0, 1).
				Bold(true)

	// Help and Footer
	StyleHelpKey = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	StyleHelpDesc = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Modal Dialogs
	StyleModalBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(1, 2)

	StyleModalTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)
)

// SignalStyle returns the appropriate styling for a given signal percent
func SignalStyle(percent int) lipgloss.Style {
	switch {
	case percent >= 65:
		return StyleSignalHigh
	case percent >= 35:
		return StyleSignalMid
	default:
		return StyleSignalLow
	}
}
