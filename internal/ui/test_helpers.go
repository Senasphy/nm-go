package ui

import (
	"github.com/Senasphy/nm-go/internal/ethernet"
	"github.com/Senasphy/nm-go/internal/iwd"
	"github.com/Senasphy/nm-go/internal/theme"
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

// SetModelEthernetData populates Ethernet devices on the model for testing
func SetModelEthernetData(m *Model, ethDevs []ethernet.Device) {
	m.ethDevices = ethDevs
	m.activeEth = nil
	for i := range m.ethDevices {
		if m.ethDevices[i].IsConnected() {
			m.activeEth = &m.ethDevices[i]
			break
		}
	}
}
