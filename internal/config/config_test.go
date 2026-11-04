package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Senasphy/nm-go/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.Theme != "tokyonight" {
		t.Errorf("expected default theme tokyonight, got %s", cfg.Theme)
	}
	if cfg.DefaultTab != "available" {
		t.Errorf("expected default tab available, got %s", cfg.DefaultTab)
	}
	if cfg.ScanIntervalSeconds != 15 {
		t.Errorf("expected default scan interval 15, got %d", cfg.ScanIntervalSeconds)
	}
}

func TestLoadNonExistent(t *testing.T) {
	tempPath := filepath.Join(t.TempDir(), "nonexistent.json")
	cfg, err := Load(tempPath)
	if err != nil {
		t.Fatalf("unexpected error loading non-existent config: %v", err)
	}
	if cfg.Theme != "tokyonight" {
		t.Errorf("expected default theme, got %s", cfg.Theme)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "nm-go", "config.json")

	cfg := &Config{
		Theme:               "gruvbox",
		DefaultTab:          "saved",
		ScanIntervalSeconds: 30,
		ColorOverrides: &ColorPalette{
			Border: "#ff0077",
			Text:   "#ffffff",
		},
		CustomThemes: []CustomThemeDefinition{
			{
				ID:   "cyber-gold",
				Name: "Cyber Gold",
				Colors: ColorPalette{
					Highlight: "#ffd700",
					Border:    "#daa520",
				},
			},
		},
	}

	if err := Save(configPath, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if loaded.Theme != "gruvbox" {
		t.Errorf("expected theme gruvbox, got %s", loaded.Theme)
	}
	if loaded.DefaultTab != "saved" {
		t.Errorf("expected default tab saved, got %s", loaded.DefaultTab)
	}
	if loaded.ScanIntervalSeconds != 30 {
		t.Errorf("expected scan interval 30, got %d", loaded.ScanIntervalSeconds)
	}
	if loaded.ColorOverrides == nil || loaded.ColorOverrides.Border != "#ff0077" {
		t.Errorf("expected color override border #ff0077, got %v", loaded.ColorOverrides)
	}
	if len(loaded.CustomThemes) != 1 || loaded.CustomThemes[0].ID != "cyber-gold" {
		t.Fatalf("expected 1 custom theme cyber-gold, got %v", loaded.CustomThemes)
	}
}

func TestApplyOverrides(t *testing.T) {
	base := theme.Default()

	palette := &ColorPalette{
		Border:    "#123456",
		Highlight: "abcdef", // Test without # prefix
	}

	th := ApplyOverrides(base, palette)

	if th.Border != lipgloss.Color("#123456") {
		t.Errorf("expected border #123456, got %v", th.Border)
	}
	if th.Highlight != lipgloss.Color("#abcdef") {
		t.Errorf("expected highlight #abcdef, got %v", th.Highlight)
	}
	// Other fields should remain unchanged from base
	if th.Text != base.Text {
		t.Errorf("expected text %v to remain untouched, got %v", base.Text, th.Text)
	}
}

func TestBuildCustomTheme(t *testing.T) {
	customDef := CustomThemeDefinition{
		ID:   "emerald",
		Name: "Emerald Forest",
		Colors: ColorPalette{
			Highlight: "#00ff88",
			Border:    "#00aa55",
			Text:      "#e0ffe0",
		},
	}

	th := customDef.BuildTheme(theme.Default())

	if th.ID != "emerald" {
		t.Errorf("expected ID emerald, got %s", th.ID)
	}
	if th.Name != "Emerald Forest" {
		t.Errorf("expected Name Emerald Forest, got %s", th.Name)
	}
	if th.Highlight != lipgloss.Color("#00ff88") {
		t.Errorf("expected highlight #00ff88, got %v", th.Highlight)
	}
	if th.Border != lipgloss.Color("#00aa55") {
		t.Errorf("expected border #00aa55, got %v", th.Border)
	}
}

func TestCorruptConfigGracefulFallback(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "bad_config.json")

	if err := os.WriteFile(configPath, []byte("{ bad json"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err == nil {
		t.Error("expected error for malformed json")
	}
	// cfg should still return safe default object
	if cfg == nil || cfg.Theme != "tokyonight" {
		t.Errorf("expected fallback to default config, got %v", cfg)
	}
}
