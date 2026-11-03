package ui

import (
	"nm-go/internal/iwd"
	"nm-go/internal/theme"
)

// SetModelData populates the model state for visual snapshots and tests
func SetModelData(
	m *Model,
	w, h int,
	nets []iwd.Network,
	known []iwd.KnownNetwork,
	devs []iwd.Device,
	st *iwd.Station,
	themePicker bool,
	searching bool,
	searchQuery string,
) {
	m.width = w
	m.height = h
	m.networks = nets
	m.knownNetworks = known
	m.devices = devs
	m.station = st
	m.isThemePicker = themePicker
	if themePicker {
		themes := theme.List()
		for i, t := range themes {
			if t.ID == m.currentTheme.ID {
				m.themePickerIndex = i
				break
			}
		}
	}
	m.filtering = searching
	if searchQuery != "" {
		m.filterInput.SetValue(searchQuery)
	}

	for i := range m.networks {
		if m.networks[i].Connected {
			m.activeNetwork = &m.networks[i]
			break
		}
	}
}
