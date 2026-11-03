# nm-go 📶

A fast, lightweight, and modern Terminal User Interface (TUI) for managing Wi-Fi connections on Linux using `iwd` (iNet wireless daemon) and the Charm ecosystem ([Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss)).

---

## ✨ Features

- **Native `iwd` Service Integration**: Communicates directly with `iwd` over the system D-Bus (`net.connman.iwd`) for blazing-fast real-time status and network discovery.
- **Available Networks Explorer**:
  - Live signal meter (`▂▄▆█`) with percent and dBm readings.
  - Security badges (Open, WPA2/WPA3 Personal PSK, 802.1x Enterprise).
  - Clear visual indicators for saved (`★ Saved`) and currently connected (`✔ Connected`) networks.
- **Saved Networks Management**:
  - View all known networks and last connected timestamps.
  - Toggle Auto-Connect per network on the fly.
  - Forget/delete network credentials with confirmation dialog.
- **Device & Adapter Diagnostics**:
  - Inspect interface name, MAC address, physical chipset model, vendor, and station operating state.
  - Toggle Wi-Fi radio power on/off with a single keypress.
- **Seamless Connecting**:
  - Direct 1-press connect to known/saved or open networks.
  - Built-in secure passphrase prompt modal for password-protected networks (with show/hide password toggle).
- **Search & Filter**: Press `/` to filter networks by SSID.
- **Auto-Refresh**: Live updates in the background every 3 seconds to keep signal levels and states current.

---

## 🛠️ Requirements

- **Linux** with `iwd` active (`systemctl is-active iwd`)
- **Go** 1.21 or later
- Optional: `iwctl` (used for initial passphrase provisioning)

---

## 🚀 Quick Start

### Build and Run

```bash
# Clone or navigate to the directory
cd nm-go

# Build the executable
go build -o nm-go .

# Run the TUI
./nm-go
```

Or run directly without building a binary:

```bash
go run .
```

---

## ⌨️ Keybindings

| Key | Action |
| :--- | :--- |
| `Tab` / `Shift+Tab` | Switch tabs (Available Networks / Saved Networks / Device Info) |
| `↑` / `k` , `↓` / `j` | Navigate list |
| `Enter` / `c` | Connect to selected network (prompts for passphrase if needed) |
| `d` | Disconnect from current network |
| `s` / `r` | Trigger a fresh Wi-Fi scan / refresh |
| `f` / `x` | Forget / remove saved network credentials |
| `a` | Toggle auto-connect for saved network |
| `p` | Toggle Wi-Fi radio power on/off |
| `/` | Search & filter networks by SSID |
| `Esc` | Clear filter / close modal |
| `?` | Toggle help modal |
| `q` / `Ctrl+C` | Quit |

---

## 📁 Architecture

The project is structured to keep code clean, modular, and unencumbered:

```text
nm-go/
├── main.go               # Application entry point, flag parsing & D-Bus check
├── internal/
│   ├── iwd/
│   │   ├── client.go      # IWD service layer: D-Bus ObjectManager & Station methods
│   │   ├── models.go      # Device, Station, Network, and KnownNetwork types
│   │   └── models_test.go # Unit tests for signal, security, and time formatting
│   └── ui/
│       ├── model.go       # Bubble Tea state model & initialization
│       ├── keys.go        # Keybinding definitions
│       ├── styles.go      # Lip Gloss themes, colors, and layout styles
│       ├── commands.go    # Asynchronous Bubble Tea commands (D-Bus dispatchers)
│       ├── update.go      # Key handling, navigation, and state transitions
│       ├── modal.go       # Password prompt and confirmation overlays
│       ├── view.go        # Full TUI layout rendering
│       └── update_test.go # Unit tests for filtering and navigation bounds
└── README.md
```

---

## 📄 License

MIT
