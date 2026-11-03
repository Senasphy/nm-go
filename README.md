# nm-go 📶

A fast, lightweight, and modern Terminal User Interface (TUI) for managing Wi-Fi networks on Linux using `iwd` (iNet wireless daemon) and the Charm ecosystem ([Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss)).

Inspired by the visual design language of [UniSocket](https://github.com/Senasphy/unisocket) and [btop](https://github.com/aristocratos/btop), `nm-go` provides a clean, responsive, keyboard-first workflow with zero clutter.

---

## ✨ Features

- **Native D-Bus + System Network Sync**:
  - Direct communication with `iwd` over the system D-Bus (`net.connman.iwd`).
  - **Full credential sync**: Recognizes networks you've already logged into via system connection managers (e.g. NetworkManager). Never re-prompts for passphrases on saved networks.
  - Accurate real-time connection state synchronized with the active hardware interface.
- **Tabbed Network Views**:
  - **Available Networks**: In-range scan results with btop-style signal meters (`■■■■■··· 65%`), security badges (`WPA2/3`, `Open`, `802.1x`), and status tags (`✔ Connected`, `★ Saved`, `● Open`).
  - **Saved Networks**: Consolidated list of all known networks, auto-connect toggles, and relative last-connected timestamps.
  - **Device Diagnostics**: Hardware interface details (`wlan0`), MAC address, operational mode, station state, and physical chipset details (`phy0`, vendor, driver).
- **Responsive Side-by-Side Theme Switcher**:
  - Press <kbd>t</kbd> to slide out a clean, compact theme selector on the right side while keeping the network table visible on the left.
  - Changes apply live in real time as you navigate through 10 curated themes (*Tokyo Night*, *Catppuccin Mocha*, *Gruvbox Dark*, *Dracula*, *Nord*, *One Dark*, *Rosé Pine*, *Monokai Pro*, *Cyberpunk Neon*, and *Retro*).
- **Uncluttered Commands Deck**:
  - Full-width bottom card with inverted keycap badges and live status alerts.
- **Interactive Search Box**:
  - Press <kbd>/</kbd> to filter SSIDs in real time.
- **Passphrase Modal**:
  - Secure modal dialog with asterisk masking and show/hide password toggle (<kbd>Ctrl+R</kbd>) for new, unconfigured networks.
- **Hardware Radio Control**:
  - Toggle Wi-Fi radio power on/off with a single keypress (<kbd>p</kbd>).

---

## 🎨 Themes Showcase

`nm-go` supports 10 built-in terminal palettes:

| Theme ID | Name | Style |
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

Press <kbd>t</kbd> inside the TUI to browse themes side-by-side, or launch directly with:

```bash
nm-go -theme gruvbox
```

---

## 🛠️ Requirements

- **Linux** with `iwd` and/or `NetworkManager` active
- **Go** 1.21 or later
- Optional: `iwctl` / `nmcli` for credential provisioning

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
sudo mv nm-go /usr/local/bin/
```

---

## 🕹️ Keyboard Shortcuts

| Key | Action |
| :--- | :--- |
| <kbd>Tab</kbd> / <kbd>Shift+Tab</kbd> | Switch tabs (*Available* / *Saved* / *Device Info*) |
| <kbd>↑</kbd> / <kbd>k</kbd> , <kbd>↓</kbd> / <kbd>j</kbd> | Navigate network list / theme selector |
| <kbd>Enter</kbd> / <kbd>c</kbd> | Connect to selected network (direct connect for known/saved) |
| <kbd>d</kbd> | Disconnect active network |
| <kbd>s</kbd> / <kbd>r</kbd> | Trigger a fresh Wi-Fi scan / refresh |
| <kbd>f</kbd> / <kbd>x</kbd> | Forget saved network credentials (with confirmation) |
| <kbd>a</kbd> | Toggle auto-connect on saved network |
| <kbd>p</kbd> | Toggle Wi-Fi power (radio on/off) |
| <kbd>/</kbd> | Open search box to filter networks by SSID |
| <kbd>t</kbd> | Toggle side-by-side theme switcher |
| <kbd>?</kbd> | Toggle help & command reference |
| <kbd>Esc</kbd> | Close theme panel / clear search / cancel action |
| <kbd>q</kbd> / <kbd>Ctrl+C</kbd> | Quit nm-go |

---

## 📄 License

MIT © [Senasphy](https://github.com/Senasphy)
