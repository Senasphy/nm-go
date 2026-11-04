package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Senasphy/nm-go/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

// ColorPalette defines individual color hex codes that can override theme colors.
type ColorPalette struct {
	Text         string `json:"text,omitempty"`
	Muted        string `json:"muted,omitempty"`
	Border       string `json:"border,omitempty"`
	Highlight    string `json:"highlight,omitempty"`
	TitleText    string `json:"title_text,omitempty"`
	Status       string `json:"status,omitempty"`
	HeaderBg     string `json:"header_bg,omitempty"`
	HeaderFg     string `json:"header_fg,omitempty"`
	HeaderBorder string `json:"header_border,omitempty"`
	SelectedBg   string `json:"selected_bg,omitempty"`
	SelectedFg   string `json:"selected_fg,omitempty"`
	Success      string `json:"success,omitempty"`
	Warning      string `json:"warning,omitempty"`
	Danger       string `json:"danger,omitempty"`
	ModalBorder  string `json:"modal_border,omitempty"`
}

// CustomThemeDefinition allows users to declare brand new themes in config.json.
type CustomThemeDefinition struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Colors ColorPalette `json:"colors"`
}

// Config represents all persistent user preferences for nm-go.
type Config struct {
	// Theme is the ID or name of the active theme (e.g. "tokyonight", "gruvbox").
	Theme string `json:"theme"`

	// DefaultTab sets the initial active tab ("available", "saved", "diagnostics").
	DefaultTab string `json:"default_tab,omitempty"`

	// ScanIntervalSeconds specifies periodic background scan frequency (0 to disable).
	ScanIntervalSeconds int `json:"scan_interval_seconds,omitempty"`

	// ColorOverrides provides optional global hex overrides applied on top of any active theme.
	ColorOverrides *ColorPalette `json:"colors,omitempty"`

	// CustomThemes allows users to define custom palettes that appear in the theme picker.
	CustomThemes []CustomThemeDefinition `json:"custom_themes,omitempty"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Theme:               "tokyonight",
		DefaultTab:          "available",
		ScanIntervalSeconds: 15,
	}
}

// DefaultConfigPath resolves the XDG config path for nm-go (~/.config/nm-go/config.json).
func DefaultConfigPath() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("unable to determine home directory: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "nm-go", "config.json"), nil
}

// Load reads and parses the JSON configuration file from path.
// If the file does not exist, it returns the default configuration without error.
func Load(path string) (*Config, error) {
	cfg := Default()

	if path == "" {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return cfg, nil
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("error reading config file: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("error parsing config JSON from %s: %w", path, err)
	}

	if strings.TrimSpace(cfg.Theme) == "" {
		cfg.Theme = "tokyonight"
	}
	if strings.TrimSpace(cfg.DefaultTab) == "" {
		cfg.DefaultTab = "available"
	}

	return cfg, nil
}

// Save atomically writes the configuration to disk formatted as clean JSON.
func Save(path string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("cannot save nil config")
	}

	if path == "" {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return err
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}
	data = append(data, '\n')

	// Atomic write via temporary file
	tmpFile := filepath.Join(dir, fmt.Sprintf(".config.tmp-%d", os.Getpid()))
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary config: %w", err)
	}

	if err := os.Rename(tmpFile, path); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to replace config file: %w", err)
	}

	return nil
}

// parseColor safely normalizes and parses a hex code or color string.
func parseColor(colorStr string, fallback lipgloss.TerminalColor) lipgloss.TerminalColor {
	trimmed := strings.TrimSpace(colorStr)
	if trimmed == "" {
		return fallback
	}
	if !strings.HasPrefix(trimmed, "#") && (len(trimmed) == 3 || len(trimmed) == 6) {
		isHex := true
		for _, c := range trimmed {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				isHex = false
				break
			}
		}
		if isHex {
			trimmed = "#" + trimmed
		}
	}
	return lipgloss.Color(trimmed)
}

// ApplyOverrides overrides colors on an existing theme using the palette values.
func ApplyOverrides(base theme.Theme, palette *ColorPalette) theme.Theme {
	if palette == nil {
		return base
	}

	th := base
	th.Text = parseColor(palette.Text, th.Text)
	th.Muted = parseColor(palette.Muted, th.Muted)
	th.Border = parseColor(palette.Border, th.Border)
	th.Highlight = parseColor(palette.Highlight, th.Highlight)
	th.TitleText = parseColor(palette.TitleText, th.TitleText)
	th.Status = parseColor(palette.Status, th.Status)
	th.HeaderBg = parseColor(palette.HeaderBg, th.HeaderBg)
	th.HeaderFg = parseColor(palette.HeaderFg, th.HeaderFg)
	th.HeaderBorder = parseColor(palette.HeaderBorder, th.HeaderBorder)
	th.SelectedBg = parseColor(palette.SelectedBg, th.SelectedBg)
	th.SelectedFg = parseColor(palette.SelectedFg, th.SelectedFg)
	th.Success = parseColor(palette.Success, th.Success)
	th.Warning = parseColor(palette.Warning, th.Warning)
	th.Danger = parseColor(palette.Danger, th.Danger)
	th.ModalBorder = parseColor(palette.ModalBorder, th.ModalBorder)

	return th
}

// BuildTheme constructs a theme.Theme instance from a CustomThemeDefinition.
func (c CustomThemeDefinition) BuildTheme(baseFallback theme.Theme) theme.Theme {
	th := baseFallback
	if strings.TrimSpace(c.ID) != "" {
		th.ID = strings.ToLower(strings.TrimSpace(c.ID))
	}
	if strings.TrimSpace(c.Name) != "" {
		th.Name = strings.TrimSpace(c.Name)
	}
	th.Description = "Custom user theme"
	return ApplyOverrides(th, &c.Colors)
}
