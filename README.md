# nm-go 📶

A fast, lightweight, and modern Terminal User Interface (TUI) for managing Wi-Fi networks on Linux using `iwd` (iNet wireless daemon) and the Charm ecosystem ([Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss)).

Inspired by the visual design language of [UniSocket](https://github.com/Senasphy/unisocket) and [btop](https://github.com/aristocratos/btop), `nm-go` blends a high-contrast aesthetic with an ergonomic, keyboard-first workflow.

---

## ✨ Features

- **Native D-Bus Service Integration**: Direct, zero-CGO communication with `iwd` over the system D-Bus (`net.connman.iwd`) via `ObjectManager` for real-time status and network discovery.
- **Tabbed Network Explorer**:
  - **Available Networks**: In-range scan results with btop-style signal meters (`■■■■■··· 65%`), security badges (`WPA2/3`, `Open`, `802.1x`), and status tags (`✔ Connected`, `★ Saved`, `● Open`).
  - **Saved Networks**: Stored network profiles, auto-connect status toggles, and relative last-connected timestamps.
  - **Device Diagnostics**: Hardware interface details (`wlan0`), MAC address, operational mode, station state, and physical chipset details (`phy0`, vendor, driver).
- **Interactive Themes**:
  - Includes 10 built-in terminal themes (*Tokyo Night*, *Catppuccin Mocha*, *Gruvbox Dark*, *Dracula*, *Nord*, *One Dark*, *Rosé Pine*, *Monokai Pro*, *Cyberpunk Neon*, and *Retro*).
  - Press <kbd>t</kbd> to open an interactive theme switcher with live preview, or start with `-theme <name>`.
- **UniSocket-Inspired Dual Bottom Deck**:
  - **Commands Card**: Inverted keycap badges for rapid command discovery.
  - **Monitor Card**: Live feed pulse indicator (`● Live feed (3s)`), active interface info, and connection status.
- **Dedicated Search Box**:
  - Press <kbd>/</kbd> to open an interactive filter input matching SSIDs in real time.
- **Passphrase Modal**:
  - Secure modal dialog with asterisk masking and show/hide password toggle (<kbd>Ctrl+R</kbd>).
- **Hardware Radio Control**:
  - Toggle Wi-Fi radio power on/off with a single keypress (<kbd>p</kbd>).

---

## 🎨 Themes Showcase

`nm-go` supports a wide array of popular palettes:

| Theme ID | Name | Style Description |
| :--- | :--- | :--- |
| `tokyonight` *(default)* | **Tokyo Night** | Downtown Tokyo neon blue and amber accents |
| `catppuccin` | **Catppuccin Mocha** | Soothing pastel dark palette with lavender highlight |
| `gruvbox` | **Gruvbox Dark** | Warm retro groove with earthy orange and yellow tones |
| `dracula` | **Dracula** | Iconic gothic dark theme with vibrant purple and green |
| `nord` | **Nord** | Arctic, north-bluish clean aesthetic |
| `onedark` | **One Dark** | Atom and Neovim balanced classic |
| `rose-pine` | **Rosé Pine** | Soho vibes, pine tones, and cozy aesthetic |
| `monokai` | **Monokai Pro** | High-contrast developer theme with bright yellow highlights |
| `cyberpunk` | **Cyberpunk Neon** | Synthwave neon pink and cyan glow |
| `default` | **Default Retro** | Classic cyan and amber terminal palette |

You can switch themes interactively by pressing <kbd>t</kbd> anytime inside the TUI, or launch with:

```bash
nm-go -theme gruvbox
```

---

## 🛠️ Requirements

- **Linux** with `iwd` daemon active (`systemctl is-active iwd`)
- **Go** 1.21 or later
- Optional: `iwctl` (used for initial WPA-PSK passphrase provisioning)

---

## 🚀 Installation

### 1. Install With Go

```bash
go install github.com/Senasphy/nm-go@latest
```

### 2. Build From Source

```bash
git clone https://github.com/Senasphy/nm-go.git
cd nm-go
go mod tidy
go build -o nm-go .
```

Move the binary to your `PATH` for system-wide access:

```bash
sudo mv nm-go /usr/local/bin/
```

---

## 🕹️ Usage & Keybindings

Launch `nm-go`:

```bash
nm-go
```

### Command-Line Flags

```text
Usage of nm-go:
  -theme string
    	Launch with specific theme (e.g. tokyonight, catppuccin, gruvbox, dracula, nord, onedark, rose-pine, monokai, cyberpunk, default) (default "tokyonight")
  -version
    	Print version and exit
  -help
    	Print help and exit
```

### Keyboard Shortcuts

| Key | Action |
| :--- | :--- |
| <kbd>Tab</kbd> / <kbd>Shift+Tab</kbd> | Switch tabs (*Available* / *Saved* / *Device Info*) |
| <kbd>↑</kbd> / <kbd>k</kbd> , <kbd>↓</kbd> / <kbd>j</kbd> | Navigate networks / options |
| <kbd>Enter</kbd> / <kbd>c</kbd> | Connect to selected network (prompts for passphrase if needed) |
| <kbd>d</kbd> | Disconnect active network |
| <kbd>s</kbd> / <kbd>r</kbd> | Trigger a fresh Wi-Fi scan / refresh |
| <kbd>f</kbd> / <kbd>x</kbd> | Forget saved network credentials (with confirmation) |
| <kbd>a</kbd> | Toggle auto-connect on saved network |
| <kbd>p</kbd> | Toggle Wi-Fi power (radio on/off) |
| <kbd>/</kbd> | Open search box to filter networks by SSID |
| <kbd>t</kbd> | Open interactive theme switcher |
| <kbd>?</kbd> | Toggle help & command reference |
| <kbd>Esc</kbd> | Close modal / clear search / cancel action |
| <kbd>q</kbd> / <kbd>Ctrl+C</kbd> | Quit nm-go |

---

## 🏗️ Architecture

`nm-go` uses a clean separation of concerns between D-Bus IPC and terminal presentation:

```text
nm-go/
├── main.go               # CLI entrypoint, flag parsing, D-Bus service check
├── internal/
│   ├── theme/
│   │   └── theme.go       # 10 curated themes & dynamic Lip Gloss style generators
│   ├── iwd/
│   │   ├── client.go      # IWD service layer: D-Bus ObjectManager & Station methods
│   │   ├── models.go      # Device, Station, Network, and KnownNetwork types
│   │   └── models_test.go # Unit tests for signal, security, and time formatting
│   └── ui/
│       ├── model.go       # Bubble Tea state model & initialization
│       ├── keys.go        # Keybinding definitions
│       ├── styles.go      # Signal meters, tabs, and keycap badges
│       ├── commands.go    # Asynchronous Bubble Tea commands (D-Bus dispatchers)
│       ├── update.go      # Event loop, keyboard handling & state transitions
│       ├── modal.go       # Password prompt, confirmation dialogs & theme switcher
│       ├── view.go        # Full TUI layout rendering (UniSocket/btop design)
│       └── update_test.go # Unit tests for filtering & cursor boundary clamping
└── README.md
```

---

## 📄 License

MIT © [Senasphy](https://github.com/Senasphy)
