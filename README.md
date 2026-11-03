# nm-go 📶

A fast, lightweight, and modern Terminal User Interface (TUI) for managing Wi-Fi networks on Linux using `iwd` (iNet wireless daemon) and the Charm ecosystem ([Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss)).

Inspired by the visual design language of [UniSocket](https://github.com/Senasphy/unisocket) and [btop](https://github.com/aristocratos/btop), `nm-go` provides a clean, responsive, keyboard-first workflow with zero clutter.

---

## ✨ Features

- **Native D-Bus Service Integration**: Direct, zero-CGO communication with `iwd` over the system D-Bus with cross-service synchronization against existing system profiles to eliminate redundant passphrase prompts.
- **Tabbed Network Management**: Dedicated views for live access point discovery, stored profile management, and physical adapter diagnostics.
- **Side-by-Side Theme Engine**: 10 curated terminal palettes with an interactive, live-preview switcher and startup flag support.
- **Keyboard-First Workflow**: Fast in-table navigation, instant SSID search filtering, hardware radio power control, and secure credential entry.

---

## 🎨 Themes Showcase

`nm-go` ships with 10 built-in palettes:

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

## 🤝 Contributing

Contributions, bug reports, and feature proposals are welcome.

### Reporting Issues

If you encounter unexpected behavior or have a feature suggestion, please open an issue on the [GitHub Issues](https://github.com/Senasphy/nm-go/issues) page. When filing a bug report, please include:
- Operating system distribution and kernel version.
- Active wireless services (`systemctl status iwd` / `NetworkManager`).
- Relevant logs or steps to reproduce the issue.

### Submitting Pull Requests

1. **Fork** the repository and create a dedicated branch:
   ```bash
   git checkout -b feat/your-feature-name
   ```
2. **Make your changes** following idiomatic Go conventions.
3. **Verify tests and code health**:
   ```bash
   go test -v ./...
   go vet ./...
   ```
4. **Commit** with clear, descriptive commit messages:
   ```bash
   git commit -m "feat(ui): add network sorting options"
   ```
5. **Push** to your fork and submit a Pull Request against `main`.

---

## 📄 License

MIT © [Senasphy](https://github.com/Senasphy)
