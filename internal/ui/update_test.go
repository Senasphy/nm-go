package ui

import (
	"path/filepath"
	"testing"

	"github.com/Senasphy/nm-go/internal/config"
	"github.com/Senasphy/nm-go/internal/iwd"
	"github.com/Senasphy/nm-go/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
)

func TestFiltering(t *testing.T) {
	m := InitialModel(nil, theme.Default(), nil, "")
	m.networks = []iwd.Network{
		{Name: "HomeWiFi", Type: "psk"},
		{Name: "Office-5G", Type: "8021x"},
		{Name: "CoffeeShop", Type: "open"},
		{Name: "Guest-Home", Type: "open"},
	}

	// Unfiltered
	if len(m.getFilteredNetworks()) != 4 {
		t.Fatalf("expected 4 networks, got %d", len(m.getFilteredNetworks()))
	}

	// Filter by "home"
	m.filterInput.SetValue("home")
	filtered := m.getFilteredNetworks()
	if len(filtered) != 2 {
		t.Fatalf("expected 2 networks matching 'home', got %d", len(filtered))
	}
	if filtered[0].Name != "HomeWiFi" || filtered[1].Name != "Guest-Home" {
		t.Errorf("unexpected filtered results: %+v", filtered)
	}

	// Filter by non-existent
	m.filterInput.SetValue("xyz999")
	if len(m.getFilteredNetworks()) != 0 {
		t.Fatalf("expected 0 networks matching 'xyz999'")
	}
}

func TestCursorClamping(t *testing.T) {
	m := InitialModel(nil, theme.Default(), nil, "")
	m.activeTab = TabAvailable
	m.networks = []iwd.Network{
		{Name: "Net1"},
		{Name: "Net2"},
	}

	m.cursorAvailable = 0
	m.navigateUp()
	if m.cursorAvailable != 0 {
		t.Errorf("expected cursor to remain at 0, got %d", m.cursorAvailable)
	}

	m.navigateDown()
	if m.cursorAvailable != 1 {
		t.Errorf("expected cursor to move to 1, got %d", m.cursorAvailable)
	}

	m.navigateDown()
	if m.cursorAvailable != 1 {
		t.Errorf("expected cursor to stay clamped at 1, got %d", m.cursorAvailable)
	}
}

func TestThemeResolution(t *testing.T) {
	th, ok := theme.Get("catppuccin")
	if !ok || th.ID != "catppuccin" {
		t.Errorf("expected catppuccin theme, got %v", th)
	}

	thGruv, ok := theme.Get("gruvbox-dark")
	if !ok || thGruv.ID != "gruvbox" {
		t.Errorf("expected gruvbox theme, got %v", thGruv)
	}
}

func TestConfigDefaultTab(t *testing.T) {
	cfg := &config.Config{
		DefaultTab: "saved",
	}
	m := InitialModel(nil, theme.Default(), cfg, "")
	if m.activeTab != TabSaved {
		t.Errorf("expected active tab TabSaved, got %v", m.activeTab)
	}

	cfgDiag := &config.Config{
		DefaultTab: "diagnostics",
	}
	mDiag := InitialModel(nil, theme.Default(), cfgDiag, "")
	if mDiag.activeTab != TabDetails {
		t.Errorf("expected active tab TabDetails, got %v", mDiag.activeTab)
	}
}

func TestThemePersistence(t *testing.T) {
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.json")
	cfg := &config.Config{
		Theme: "tokyonight",
	}

	m := InitialModel(nil, theme.Default(), cfg, cfgPath)
	m.isThemePicker = true
	m.themePickerIndex = 0

	// Navigate down in theme picker
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updatedM.(Model)

	// Close theme picker and persist
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedM.(Model)

	if m.isThemePicker {
		t.Error("expected theme picker to be closed")
	}
	if m.cfg.Theme == "tokyonight" {
		t.Error("expected theme to have changed in cfg")
	}
}
