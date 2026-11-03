package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme encapsulates the color palette and styles for nm-go.
type Theme struct {
	ID          string
	Name        string
	Description string

	// Base palette
	Text        lipgloss.TerminalColor
	Muted       lipgloss.TerminalColor
	Border      lipgloss.TerminalColor
	Highlight   lipgloss.TerminalColor
	TitleText   lipgloss.TerminalColor
	Status      lipgloss.TerminalColor

	// Table styles
	HeaderBg     lipgloss.TerminalColor
	HeaderFg     lipgloss.TerminalColor
	HeaderBorder lipgloss.TerminalColor
	SelectedBg   lipgloss.TerminalColor
	SelectedFg   lipgloss.TerminalColor

	// Accents
	Success     lipgloss.TerminalColor
	Warning     lipgloss.TerminalColor
	Danger      lipgloss.TerminalColor
	ModalBorder lipgloss.TerminalColor
}

func (t Theme) TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(t.TitleText).
		Background(t.Highlight).
		Padding(0, 1)
}

func (t Theme) TableContainerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Foreground(t.Text).
		Padding(0, 1)
}

func (t Theme) CardStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Foreground(t.Text).
		Padding(0, 1)
}

func (t Theme) ModalStyle() lipgloss.Style {
	border := t.Border
	if border == nil {
		border = t.ModalBorder
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Foreground(t.Text).
		Padding(1, 2)
}

func (t Theme) HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(t.HeaderFg).
		Background(t.HeaderBg).
		Padding(0, 1)
}

func (t Theme) SelectedRowStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(t.SelectedFg).
		Background(t.SelectedBg).
		Padding(0, 1)
}

func (t Theme) RowStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Text).
		Padding(0, 1)
}

var allThemes = []Theme{
	{
		ID:          "tokyonight",
		Name:        "Tokyo Night",
		Description: "Downtown Tokyo night neon lights",
		Text:        lipgloss.Color("#c0caf5"),
		Muted:       lipgloss.Color("#565f89"),
		Border:      lipgloss.Color("#7aa2f7"),
		Highlight:   lipgloss.Color("#7aa2f7"),
		TitleText:   lipgloss.Color("#16161e"),
		Status:      lipgloss.Color("#e0af68"),
		HeaderBg:    lipgloss.Color("#24283b"),
		HeaderFg:    lipgloss.Color("#7dcfff"),
		HeaderBorder: lipgloss.Color("#7aa2f7"),
		SelectedBg:  lipgloss.Color("#7aa2f7"),
		SelectedFg:  lipgloss.Color("#16161e"),
		Success:     lipgloss.Color("#9ece6a"),
		Warning:     lipgloss.Color("#e0af68"),
		Danger:      lipgloss.Color("#f7768e"),
		ModalBorder: lipgloss.Color("#7aa2f7"),
	},
	{
		ID:          "catppuccin",
		Name:        "Catppuccin Mocha",
		Description: "Soothing pastel dark palette",
		Text:        lipgloss.Color("#cdd6f4"),
		Muted:       lipgloss.Color("#6c7086"),
		Border:      lipgloss.Color("#89b4fa"),
		Highlight:   lipgloss.Color("#cba6f7"),
		TitleText:   lipgloss.Color("#11111b"),
		Status:      lipgloss.Color("#fab387"),
		HeaderBg:    lipgloss.Color("#313244"),
		HeaderFg:    lipgloss.Color("#cba6f7"),
		HeaderBorder: lipgloss.Color("#89b4fa"),
		SelectedBg:  lipgloss.Color("#cba6f7"),
		SelectedFg:  lipgloss.Color("#11111b"),
		Success:     lipgloss.Color("#a6e3a1"),
		Warning:     lipgloss.Color("#f9e2af"),
		Danger:      lipgloss.Color("#f38ba8"),
		ModalBorder: lipgloss.Color("#89b4fa"),
	},
	{
		ID:          "gruvbox",
		Name:        "Gruvbox Dark",
		Description: "Warm retro groove with earthy tones",
		Text:        lipgloss.Color("#ebdbb2"),
		Muted:       lipgloss.Color("#928374"),
		Border:      lipgloss.Color("#fe8019"),
		Highlight:   lipgloss.Color("#fe8019"),
		TitleText:   lipgloss.Color("#282828"),
		Status:      lipgloss.Color("#fabd2f"),
		HeaderBg:    lipgloss.Color("#3c3836"),
		HeaderFg:    lipgloss.Color("#fabd2f"),
		HeaderBorder: lipgloss.Color("#fe8019"),
		SelectedBg:  lipgloss.Color("#fe8019"),
		SelectedFg:  lipgloss.Color("#282828"),
		Success:     lipgloss.Color("#b8bb26"),
		Warning:     lipgloss.Color("#fabd2f"),
		Danger:      lipgloss.Color("#fb4934"),
		ModalBorder: lipgloss.Color("#fe8019"),
	},
	{
		ID:          "dracula",
		Name:        "Dracula",
		Description: "Iconic gothic dark theme with vibrant accents",
		Text:        lipgloss.Color("#f8f8f2"),
		Muted:       lipgloss.Color("#6272a4"),
		Border:      lipgloss.Color("#bd93f9"),
		Highlight:   lipgloss.Color("#bd93f9"),
		TitleText:   lipgloss.Color("#282a36"),
		Status:      lipgloss.Color("#ffb86c"),
		HeaderBg:    lipgloss.Color("#44475a"),
		HeaderFg:    lipgloss.Color("#50fa7b"),
		HeaderBorder: lipgloss.Color("#bd93f9"),
		SelectedBg:  lipgloss.Color("#bd93f9"),
		SelectedFg:  lipgloss.Color("#282a36"),
		Success:     lipgloss.Color("#50fa7b"),
		Warning:     lipgloss.Color("#ffb86c"),
		Danger:      lipgloss.Color("#ff5555"),
		ModalBorder: lipgloss.Color("#bd93f9"),
	},
	{
		ID:          "nord",
		Name:        "Nord",
		Description: "Arctic, north-bluish clean palette",
		Text:        lipgloss.Color("#eceff4"),
		Muted:       lipgloss.Color("#4c566a"),
		Border:      lipgloss.Color("#88c0d0"),
		Highlight:   lipgloss.Color("#88c0d0"),
		TitleText:   lipgloss.Color("#2e3440"),
		Status:      lipgloss.Color("#ebcb8b"),
		HeaderBg:    lipgloss.Color("#3b4252"),
		HeaderFg:    lipgloss.Color("#8fbcbb"),
		HeaderBorder: lipgloss.Color("#81a1c1"),
		SelectedBg:  lipgloss.Color("#88c0d0"),
		SelectedFg:  lipgloss.Color("#2e3440"),
		Success:     lipgloss.Color("#a3be8c"),
		Warning:     lipgloss.Color("#ebcb8b"),
		Danger:      lipgloss.Color("#bf616a"),
		ModalBorder: lipgloss.Color("#88c0d0"),
	},
	{
		ID:          "onedark",
		Name:        "One Dark",
		Description: "Atom & Neovim balanced classic",
		Text:        lipgloss.Color("#abb2bf"),
		Muted:       lipgloss.Color("#5c6370"),
		Border:      lipgloss.Color("#61afef"),
		Highlight:   lipgloss.Color("#61afef"),
		TitleText:   lipgloss.Color("#282c34"),
		Status:      lipgloss.Color("#e5c07b"),
		HeaderBg:    lipgloss.Color("#353b45"),
		HeaderFg:    lipgloss.Color("#98c379"),
		HeaderBorder: lipgloss.Color("#61afef"),
		SelectedBg:  lipgloss.Color("#61afef"),
		SelectedFg:  lipgloss.Color("#282c34"),
		Success:     lipgloss.Color("#98c379"),
		Warning:     lipgloss.Color("#e5c07b"),
		Danger:      lipgloss.Color("#e06c75"),
		ModalBorder: lipgloss.Color("#61afef"),
	},
	{
		ID:          "rose-pine",
		Name:        "Rosé Pine",
		Description: "Soho vibes, pine tones, and cozy aesthetic",
		Text:        lipgloss.Color("#e0def4"),
		Muted:       lipgloss.Color("#6e6a86"),
		Border:      lipgloss.Color("#ebbcba"),
		Highlight:   lipgloss.Color("#ebbcba"),
		TitleText:   lipgloss.Color("#191724"),
		Status:      lipgloss.Color("#f6c177"),
		HeaderBg:    lipgloss.Color("#26233a"),
		HeaderFg:    lipgloss.Color("#9ccfd8"),
		HeaderBorder: lipgloss.Color("#ebbcba"),
		SelectedBg:  lipgloss.Color("#ebbcba"),
		SelectedFg:  lipgloss.Color("#191724"),
		Success:     lipgloss.Color("#9ccfd8"),
		Warning:     lipgloss.Color("#f6c177"),
		Danger:      lipgloss.Color("#eb6f92"),
		ModalBorder: lipgloss.Color("#ebbcba"),
	},
	{
		ID:          "monokai",
		Name:        "Monokai Pro",
		Description: "High-contrast vibrant developer theme",
		Text:        lipgloss.Color("#fcfcfa"),
		Muted:       lipgloss.Color("#727072"),
		Border:      lipgloss.Color("#ffd866"),
		Highlight:   lipgloss.Color("#ffd866"),
		TitleText:   lipgloss.Color("#2d2a2e"),
		Status:      lipgloss.Color("#fc9867"),
		HeaderBg:    lipgloss.Color("#403e41"),
		HeaderFg:    lipgloss.Color("#a9dc76"),
		HeaderBorder: lipgloss.Color("#ffd866"),
		SelectedBg:  lipgloss.Color("#ffd866"),
		SelectedFg:  lipgloss.Color("#2d2a2e"),
		Success:     lipgloss.Color("#a9dc76"),
		Warning:     lipgloss.Color("#ffd866"),
		Danger:      lipgloss.Color("#ff6188"),
		ModalBorder: lipgloss.Color("#ff6188"),
	},
	{
		ID:          "cyberpunk",
		Name:        "Cyberpunk Neon",
		Description: "Neon synthwave glow for night owls",
		Text:        lipgloss.Color("#ffffff"),
		Muted:       lipgloss.Color("#846b8a"),
		Border:      lipgloss.Color("#00ffff"),
		Highlight:   lipgloss.Color("#ff007f"),
		TitleText:   lipgloss.Color("#ffffff"),
		Status:      lipgloss.Color("#ffe600"),
		HeaderBg:    lipgloss.Color("#3f1052"),
		HeaderFg:    lipgloss.Color("#ffe600"),
		HeaderBorder: lipgloss.Color("#00ffff"),
		SelectedBg:  lipgloss.Color("#ff007f"),
		SelectedFg:  lipgloss.Color("#ffffff"),
		Success:     lipgloss.Color("#00ff66"),
		Warning:     lipgloss.Color("#ffe600"),
		Danger:      lipgloss.Color("#ff0055"),
		ModalBorder: lipgloss.Color("#00ffff"),
	},
	{
		ID:          "default",
		Name:        "Default Retro",
		Description: "Classic cyan & amber terminal palette",
		Text:        lipgloss.Color("230"),
		Muted:       lipgloss.Color("245"),
		Border:      lipgloss.Color("31"),
		Highlight:   lipgloss.Color("31"),
		TitleText:   lipgloss.Color("230"),
		Status:      lipgloss.Color("222"),
		HeaderBg:    lipgloss.Color("24"),
		HeaderFg:    lipgloss.Color("229"),
		HeaderBorder: lipgloss.Color("63"),
		SelectedBg:  lipgloss.Color("31"),
		SelectedFg:  lipgloss.Color("230"),
		Success:     lipgloss.Color("78"),
		Warning:     lipgloss.Color("222"),
		Danger:      lipgloss.Color("196"),
		ModalBorder: lipgloss.Color("31"),
	},
}

// List returns a slice of all available themes.
func List() []Theme {
	themes := make([]Theme, len(allThemes))
	copy(themes, allThemes)
	return themes
}

// Default returns the default theme (Tokyo Night).
func Default() Theme {
	return allThemes[0]
}

// Get resolves a theme by ID, name, or common aliases.
func Get(query string) (Theme, bool) {
	clean := strings.ToLower(strings.TrimSpace(query))
	clean = strings.ReplaceAll(clean, "_", "-")
	clean = strings.ReplaceAll(clean, " ", "-")

	for _, th := range allThemes {
		if th.ID == clean || strings.ToLower(th.Name) == clean {
			return th, true
		}
	}

	switch clean {
	case "tokyo", "tokyo-night":
		return Get("tokyonight")
	case "mocha", "catppuccin-mocha":
		return Get("catppuccin")
	case "gruv", "gruvbox-dark":
		return Get("gruvbox")
	case "one-dark", "atom":
		return Get("onedark")
	case "rose", "rosepine":
		return Get("rose-pine")
	case "monokaipro", "monokai-pro":
		return Get("monokai")
	case "synthwave", "neon":
		return Get("cyberpunk")
	case "retro":
		return Get("default")
	}

	for _, th := range allThemes {
		if strings.HasPrefix(th.ID, clean) || strings.HasPrefix(strings.ToLower(th.Name), clean) {
			return th, true
		}
	}

	return Default(), false
}
