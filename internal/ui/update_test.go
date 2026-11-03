package ui

import (
	"testing"

	"nm-go/internal/iwd"
)

func TestFiltering(t *testing.T) {
	m := InitialModel(nil)
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
	m := InitialModel(nil)
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
