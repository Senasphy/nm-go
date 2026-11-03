package iwd

import (
	"testing"
	"time"
)

func TestSignalCalculations(t *testing.T) {
	tests := []struct {
		name        string
		signal      int16
		expectedDBm float64
		minPct      int
		maxPct      int
		expectedBar string
	}{
		{
			name:        "Strong signal (-40 dBm)",
			signal:      -4000,
			expectedDBm: -40.0,
			minPct:      95,
			maxPct:      100,
			expectedBar: "▂▄▆█",
		},
		{
			name:        "Medium signal (-65 dBm)",
			signal:      -6500,
			expectedDBm: -65.0,
			minPct:      55,
			maxPct:      65,
			expectedBar: "▂▄▆_",
		},
		{
			name:        "Weak signal (-85 dBm)",
			signal:      -8500,
			expectedDBm: -85.0,
			minPct:      15,
			maxPct:      25,
			expectedBar: "▂___",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			net := Network{Signal: tc.signal}
			dbm := net.SignalDBm()
			if dbm != tc.expectedDBm {
				t.Errorf("expected %.2f dBm, got %.2f", tc.expectedDBm, dbm)
			}
			pct := net.SignalPercent()
			if pct < tc.minPct || pct > tc.maxPct {
				t.Errorf("expected pct between %d and %d, got %d", tc.minPct, tc.maxPct, pct)
			}
			bars := net.SignalBars()
			if bars != tc.expectedBar {
				t.Errorf("expected bars %s, got %s", tc.expectedBar, bars)
			}
		})
	}
}

func TestSecurityLabel(t *testing.T) {
	if (Network{Type: "open"}).SecurityLabel() != "Open" {
		t.Errorf("expected 'Open'")
	}
	if (Network{Type: "psk"}).SecurityLabel() != "WPA2/WPA3 (PSK)" {
		t.Errorf("expected 'WPA2/WPA3 (PSK)'")
	}
	if (Network{Type: "8021x"}).SecurityLabel() != "Enterprise (802.1x)" {
		t.Errorf("expected 'Enterprise (802.1x)'")
	}
}

func TestFormattedLastConnected(t *testing.T) {
	knEmpty := KnownNetwork{}
	if knEmpty.FormattedLastConnected() != "Never" {
		t.Errorf("expected 'Never', got %s", knEmpty.FormattedLastConnected())
	}

	recent := time.Now().Add(-5 * time.Minute).Format(time.RFC3339)
	knRecent := KnownNetwork{LastConnectedTime: recent}
	res := knRecent.FormattedLastConnected()
	if res != "5m ago" {
		t.Errorf("expected '5m ago', got %s", res)
	}
}
