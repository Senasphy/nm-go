package iwd

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetSystemSavedConnections returns SSIDs saved by system connection managers (e.g. NetworkManager)
func GetSystemSavedConnections() []string {
	cmd := exec.Command("nmcli", "-t", "-f", "NAME,TYPE", "connection", "show")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var saved []string
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) >= 2 {
			name := parts[0]
			connType := parts[1]
			if strings.Contains(connType, "wireless") || strings.Contains(connType, "wifi") || strings.Contains(connType, "802-11") {
				saved = append(saved, name)
			}
		}
	}
	return saved
}

// GetSystemActiveWirelessConnection returns the SSID and true if wlan is actively connected via system network manager
func GetSystemActiveWirelessConnection(devName string) (string, bool) {
	cmd := exec.Command("nmcli", "-t", "-f", "NAME,TYPE,DEVICE", "connection", "show", "--active")
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			name := parts[0]
			connType := parts[1]
			dev := parts[2]
			if (devName == "" || dev == devName) && (strings.Contains(connType, "wireless") || strings.Contains(connType, "wifi") || strings.Contains(connType, "802-11")) {
				return name, true
			}
		}
	}
	return "", false
}

// ConnectSystemSaved connects to an existing system-saved connection profile without re-prompting for credentials
func ConnectSystemSaved(name string) error {
	cmd := exec.Command("nmcli", "connection", "up", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		outputMsg := strings.TrimSpace(string(out))
		if outputMsg != "" {
			return fmt.Errorf("%s", outputMsg)
		}
		return err
	}
	return nil
}

// DisconnectSystem disconnects the wireless interface
func DisconnectSystem(devName string) error {
	cmd := exec.Command("nmcli", "device", "disconnect", devName)
	return cmd.Run()
}
