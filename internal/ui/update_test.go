package ui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Senasphy/nm-go/internal/config"
	"github.com/Senasphy/nm-go/internal/ethernet"
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

func TestEthernetFiltering(t *testing.T) {
	m := InitialModel(nil, theme.Default(), nil, "")
	m.ethDevices = []ethernet.Device{
		{Name: "enp0s31f6", Connection: "Wired connection 1", IP4: []string{"192.168.1.50/24"}, Driver: "e1000e"},
		{Name: "eth0", Connection: "Office LAN", IP4: []string{"10.0.0.12/24"}, Driver: "r8169"},
		{Name: "usb0", Connection: "", IP4: nil, Driver: "cdc_ether"},
	}

	// Unfiltered
	if len(m.getFilteredEthDevices()) != 3 {
		t.Fatalf("expected 3 eth devices, got %d", len(m.getFilteredEthDevices()))
	}

	// Filter by IP "192.168"
	m.filterInput.SetValue("192.168")
	filtered := m.getFilteredEthDevices()
	if len(filtered) != 1 || filtered[0].Name != "enp0s31f6" {
		t.Fatalf("expected 1 result matching '192.168', got %+v", filtered)
	}

	// Filter by driver "r8169"
	m.filterInput.SetValue("r8169")
	filtered = m.getFilteredEthDevices()
	if len(filtered) != 1 || filtered[0].Name != "eth0" {
		t.Fatalf("expected 1 result matching 'r8169', got %+v", filtered)
	}

	// Filter by connection name "Office"
	m.filterInput.SetValue("Office")
	filtered = m.getFilteredEthDevices()
	if len(filtered) != 1 || filtered[0].Name != "eth0" {
		t.Fatalf("expected 1 result matching 'Office', got %+v", filtered)
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

func TestEthernetCursorClamping(t *testing.T) {
	m := InitialModel(nil, theme.Default(), nil, "")
	m.activeTab = TabEthernet
	m.ethDevices = []ethernet.Device{
		{Name: "eth0"},
		{Name: "eth1"},
	}

	m.cursorEthernet = 0
	m.navigateUp()
	if m.cursorEthernet != 0 {
		t.Errorf("expected cursor to remain at 0, got %d", m.cursorEthernet)
	}

	m.navigateDown()
	if m.cursorEthernet != 1 {
		t.Errorf("expected cursor to move to 1, got %d", m.cursorEthernet)
	}

	m.navigateDown()
	if m.cursorEthernet != 1 {
		t.Errorf("expected cursor to stay clamped at 1, got %d", m.cursorEthernet)
	}
}

func TestTabCycling(t *testing.T) {
	m := InitialModel(nil, theme.Default(), nil, "")
	if m.activeTab != TabAvailable {
		t.Fatalf("initial active tab should be TabAvailable, got %v", m.activeTab)
	}

	// Press Tab: Available -> Saved
	updatedM, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedM.(Model)
	if m.activeTab != TabSaved {
		t.Errorf("expected TabSaved, got %v", m.activeTab)
	}

	// Press Tab: Saved -> Ethernet
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedM.(Model)
	if m.activeTab != TabEthernet {
		t.Errorf("expected TabEthernet, got %v", m.activeTab)
	}

	// Press Tab: Ethernet -> Diagnostics
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedM.(Model)
	if m.activeTab != TabDetails {
		t.Errorf("expected TabDetails, got %v", m.activeTab)
	}

	// Press Tab: Diagnostics -> Available
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updatedM.(Model)
	if m.activeTab != TabAvailable {
		t.Errorf("expected TabAvailable, got %v", m.activeTab)
	}

	// Press Shift+Tab: Available -> Diagnostics
	updatedM, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updatedM.(Model)
	if m.activeTab != TabDetails {
		t.Errorf("expected TabDetails on Shift+Tab, got %v", m.activeTab)
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

	cfgEth := &config.Config{
		DefaultTab: "ethernet",
	}
	mEth := InitialModel(nil, theme.Default(), cfgEth, "")
	if mEth.activeTab != TabEthernet {
		t.Errorf("expected active tab TabEthernet, got %v", mEth.activeTab)
	}

	cfgWired := &config.Config{
		DefaultTab: "wired",
	}
	mWired := InitialModel(nil, theme.Default(), cfgWired, "")
	if mWired.activeTab != TabEthernet {
		t.Errorf("expected active tab TabEthernet for 'wired', got %v", mWired.activeTab)
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

func TestEthernetViewRendering(t *testing.T) {
	m := InitialModel(nil, theme.Default(), nil, "")
	m.width = 100
	m.height = 30
	m.activeTab = TabEthernet
	m.ethDevices = []ethernet.Device{
		{
			Name:       "enp0s31f6",
			Carrier:    true,
			OperState:  "up",
			Speed:      1000,
			Duplex:     "full",
			IP4:        []string{"192.168.1.150/24"},
			Driver:     "e1000e",
			Connection: "Wired connection 1",
			IsUp:       true,
		},
	}

	view := m.View()
	if !strings.Contains(view, "Ethernet (1)") {
		t.Errorf("expected View to contain 'Ethernet (1)', got: %s", view)
	}
	if !strings.Contains(view, "enp0s31f6") {
		t.Errorf("expected View to contain 'enp0s31f6', got: %s", view)
	}
	if !strings.Contains(view, "192.168.1.150/24") {
		t.Errorf("expected View to contain IP, got: %s", view)
	}
	if !strings.Contains(view, "1 Gbps (Full)") {
		t.Errorf("expected View to contain '1 Gbps (Full)', got: %s", view)
	}
}
