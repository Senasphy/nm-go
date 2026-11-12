package ethernet

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeviceMethods(t *testing.T) {
	d := Device{
		Name:      "eth0",
		MAC:       "aa:bb:cc:dd:ee:ff",
		State:     "connected",
		Carrier:   true,
		OperState: "up",
		Speed:     1000,
		Duplex:    "full",
		IP4:       []string{"192.168.1.100/24"},
		IP6:       []string{"fe80::1/64"},
		Gateway:   "192.168.1.1",
	}

	if !d.IsConnected() {
		t.Errorf("expected device to be connected")
	}

	if d.PrimaryIP() != "192.168.1.100/24" {
		t.Errorf("unexpected primary IP: %s", d.PrimaryIP())
	}

	if d.PrimaryIPClean() != "192.168.1.100" {
		t.Errorf("unexpected clean primary IP: %s", d.PrimaryIPClean())
	}

	if d.FormattedSpeed() != "1 Gbps (Full)" {
		t.Errorf("unexpected formatted speed: %s", d.FormattedSpeed())
	}

	if d.StatusLabel() != "Connected" {
		t.Errorf("unexpected status label: %s", d.StatusLabel())
	}
}

func TestDeviceSpeedFormatting(t *testing.T) {
	tests := []struct {
		carrier  bool
		speed    int
		duplex   string
		expected string
	}{
		{false, 1000, "full", "-"},
		{true, 0, "full", "-"},
		{true, -1, "full", "-"},
		{true, 100, "full", "100 Mbps (Full)"},
		{true, 100, "half", "100 Mbps (Half)"},
		{true, 1000, "full", "1 Gbps (Full)"},
		{true, 2500, "full", "2.5 Gbps (Full)"},
		{true, 10000, "full", "10 Gbps (Full)"},
		{true, 100, "unknown", "100 Mbps"},
	}

	for _, tt := range tests {
		dev := Device{
			Carrier: tt.carrier,
			Speed:   tt.speed,
			Duplex:  tt.duplex,
		}
		res := dev.FormattedSpeed()
		if res != tt.expected {
			t.Errorf("FormattedSpeed() with carrier=%v, speed=%d, duplex=%s: got %s, want %s",
				tt.carrier, tt.speed, tt.duplex, res, tt.expected)
		}
	}
}

func TestDeviceStatusLabel(t *testing.T) {
	tests := []struct {
		dev      Device
		expected string
	}{
		{
			dev:      Device{State: "connected"},
			expected: "Connected",
		},
		{
			dev:      Device{Carrier: true, IP4: []string{"10.0.0.5/24"}},
			expected: "Connected",
		},
		{
			dev:      Device{Carrier: true, IP4: nil},
			expected: "Link Up (No IP)",
		},
		{
			dev:      Device{State: "connecting"},
			expected: "Connecting",
		},
		{
			dev:      Device{OperState: "down", IsUp: false},
			expected: "Disabled / Down",
		},
		{
			dev:      Device{State: "unavailable"},
			expected: "Cable Unplugged",
		},
		{
			dev:      Device{Carrier: false, IsUp: true, OperState: "up"},
			expected: "Unplugged",
		},
	}

	for _, tt := range tests {
		res := tt.dev.StatusLabel()
		if res != tt.expected {
			t.Errorf("StatusLabel() for %+v: got %s, want %s", tt.dev, res, tt.expected)
		}
	}
}

func TestParseHexIPv4(t *testing.T) {
	// Little endian: 019B1C0A -> 10.28.155.1
	ip := parseHexIPv4("019B1C0A")
	if ip != "10.28.155.1" {
		t.Errorf("expected 10.28.155.1, got %s", ip)
	}

	// 0101A8C0 -> 192.168.1.1
	ip2 := parseHexIPv4("0101A8C0")
	if ip2 != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %s", ip2)
	}

	// Invalid length
	if parseHexIPv4("123") != "" {
		t.Errorf("expected empty string for invalid hex length")
	}
}

func TestGetDefaultGateways(t *testing.T) {
	tmpDir := t.TempDir()
	routeFile := filepath.Join(tmpDir, "route")
	content := "Iface\tDestination\tGateway\tFlags\tRefCnt\tUse\tMetric\tMask\tMTU\tWindow\tIRTT\n" +
		"eth0\t00000000\t0101A8C0\t0003\t0\t0\t100\t00000000\t0\t0\t0\n" +
		"wlan0\t00000000\t019B1C0A\t0003\t0\t0\t600\t00000000\t0\t0\t0\n" +
		"eth0\t0001A8C0\t00000000\t0001\t0\t0\t100\t00FFFFFF\t0\t0\t0\n"

	if err := os.WriteFile(routeFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write mock route file: %v", err)
	}

	gwMap := getDefaultGateways(routeFile)
	if gwMap["eth0"] != "192.168.1.1" {
		t.Errorf("expected eth0 gateway 192.168.1.1, got %s", gwMap["eth0"])
	}
	if gwMap["wlan0"] != "10.28.155.1" {
		t.Errorf("expected wlan0 gateway 10.28.155.1, got %s", gwMap["wlan0"])
	}
}

func TestGetDevicesFromSysfsMock(t *testing.T) {
	tmpSysfs := t.TempDir()

	// 1. Mock loopback 'lo'
	loDir := filepath.Join(tmpSysfs, "lo")
	_ = os.MkdirAll(loDir, 0755)
	_ = os.WriteFile(filepath.Join(loDir, "type"), []byte("772\n"), 0644)

	// 2. Mock wifi 'wlan0'
	wifiDir := filepath.Join(tmpSysfs, "wlan0")
	_ = os.MkdirAll(wifiDir, 0755)
	_ = os.MkdirAll(filepath.Join(wifiDir, "wireless"), 0755)
	_ = os.WriteFile(filepath.Join(wifiDir, "type"), []byte("1\n"), 0644)

	// 3. Mock ethernet 'eth0'
	ethDir := filepath.Join(tmpSysfs, "eth0")
	_ = os.MkdirAll(ethDir, 0755)
	_ = os.MkdirAll(filepath.Join(ethDir, "device"), 0755)
	_ = os.MkdirAll(filepath.Join(ethDir, "statistics"), 0755)
	_ = os.WriteFile(filepath.Join(ethDir, "type"), []byte("1\n"), 0644)
	_ = os.WriteFile(filepath.Join(ethDir, "address"), []byte("11:22:33:44:55:66\n"), 0644)
	_ = os.WriteFile(filepath.Join(ethDir, "carrier"), []byte("1\n"), 0644)
	_ = os.WriteFile(filepath.Join(ethDir, "operstate"), []byte("up\n"), 0644)
	_ = os.WriteFile(filepath.Join(ethDir, "speed"), []byte("1000\n"), 0644)
	_ = os.WriteFile(filepath.Join(ethDir, "duplex"), []byte("full\n"), 0644)
	_ = os.WriteFile(filepath.Join(ethDir, "mtu"), []byte("1500\n"), 0644)
	_ = os.WriteFile(filepath.Join(ethDir, "statistics", "rx_bytes"), []byte("102400\n"), 0644)
	_ = os.WriteFile(filepath.Join(ethDir, "statistics", "tx_bytes"), []byte("204800\n"), 0644)

	// 4. Mock docker0 virtual bridge
	dockDir := filepath.Join(tmpSysfs, "docker0")
	_ = os.MkdirAll(dockDir, 0755)
	_ = os.WriteFile(filepath.Join(dockDir, "type"), []byte("1\n"), 0644)

	devices, err := getDevicesFromSysfs(tmpSysfs)
	if err != nil {
		t.Fatalf("getDevicesFromSysfs returned error: %v", err)
	}

	if len(devices) != 1 {
		t.Fatalf("expected 1 ethernet device, got %d", len(devices))
	}

	dev := devices[0]
	if dev.Name != "eth0" {
		t.Errorf("expected device eth0, got %s", dev.Name)
	}
	if dev.MAC != "11:22:33:44:55:66" {
		t.Errorf("expected MAC 11:22:33:44:55:66, got %s", dev.MAC)
	}
	if !dev.Carrier {
		t.Errorf("expected Carrier true")
	}
	if dev.Speed != 1000 {
		t.Errorf("expected Speed 1000, got %d", dev.Speed)
	}
	if dev.Duplex != "full" {
		t.Errorf("expected Duplex full, got %s", dev.Duplex)
	}
	if dev.RxBytes != 102400 || dev.TxBytes != 204800 {
		t.Errorf("unexpected rx/tx bytes: %d/%d", dev.RxBytes, dev.TxBytes)
	}
}

func TestLiveSystemGetDevices(t *testing.T) {
	client := NewClient()
	devs, err := client.GetDevices()
	if err != nil {
		t.Logf("GetDevices error on this environment: %v", err)
		return
	}
	t.Logf("Discovered %d ethernet devices on live system", len(devs))
	for _, d := range devs {
		t.Logf("Device: %+v", d)
	}
}
