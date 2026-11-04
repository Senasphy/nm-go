package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRegisterCustomTheme(t *testing.T) {
	initialCount := len(List())

	custom := Theme{
		ID:        "sunset",
		Name:      "Sunset",
		Highlight: lipgloss.Color("#ff5e3a"),
	}

	Register(custom)

	if len(List()) != initialCount+1 {
		t.Errorf("expected theme count %d, got %d", initialCount+1, len(List()))
	}

	found, ok := Get("sunset")
	if !ok {
		t.Fatal("expected to find registered theme 'sunset'")
	}
	if found.Name != "Sunset" {
		t.Errorf("expected Name 'Sunset', got %s", found.Name)
	}

	// Test case-insensitivity and alias resolution
	foundAlias, ok := Get("SUNSET")
	if !ok || foundAlias.ID != "sunset" {
		t.Errorf("expected case-insensitive resolution for SUNSET")
	}
}
