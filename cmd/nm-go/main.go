package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Senasphy/nm-go/internal/config"
	"github.com/Senasphy/nm-go/internal/iwd"
	"github.com/Senasphy/nm-go/internal/theme"
	"github.com/Senasphy/nm-go/internal/ui"
)

const version = "0.3.0"

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	showHelp := flag.Bool("help", false, "Print help and exit")
	themeName := flag.String("theme", "", "Launch with specific theme override (e.g. tokyonight, catppuccin, gruvbox, dracula, nord, onedark, rose-pine, monokai, cyberpunk, default)")
	configPathFlag := flag.String("config", "", "Custom path to config.json (default: ~/.config/nm-go/config.json)")
	dumpConfig := flag.Bool("dump-config", false, "Print a starter configuration template in JSON to stdout and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("nm-go v%s (Wi-Fi TUI for iwd)\n", version)
		os.Exit(0)
	}

	if *dumpConfig {
		sample := `{
  "theme": "tokyonight",
  "default_tab": "available",
  "scan_interval_seconds": 15,
  "colors": {
    "border": "#29a4bd",
    "highlight": "#7aa2f7",
    "selected_bg": "#7aa2f7",
    "selected_fg": "#15161e",
    "text": "#c0caf5",
    "muted": "#565f89",
    "success": "#9ece6a",
    "warning": "#e0af68",
    "danger": "#f7768e"
  },
  "custom_themes": [
    {
      "id": "emerald",
      "name": "Emerald Forest",
      "colors": {
        "text": "#e6edf3",
        "muted": "#7d8590",
        "border": "#238636",
        "highlight": "#2ea043",
        "title_text": "#0d1117",
        "status": "#56d364",
        "header_bg": "#161b22",
        "header_fg": "#3fb950",
        "header_border": "#238636",
        "selected_bg": "#2ea043",
        "selected_fg": "#0d1117",
        "success": "#3fb950",
        "warning": "#d29922",
        "danger": "#f85149",
        "modal_border": "#2ea043"
      }
    }
  ]
}`
		fmt.Println(sample)
		os.Exit(0)
	}

	if *showHelp {
		fmt.Println("nm-go: A clean terminal UI for managing Wi-Fi networks using iwd")
		fmt.Println("\nUsage:")
		fmt.Println("  nm-go [flags]")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Verify that iwd is running on the system bus
	if err := iwd.CheckIwdService(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please ensure iwd is running: sudo systemctl start iwd\n")
		os.Exit(1)
	}

	client, err := iwd.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to iwd: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Determine configuration file path
	configPath := *configPathFlag
	if configPath == "" {
		var err error
		configPath, err = config.DefaultConfigPath()
		if err != nil {
			configPath = ""
		}
	}

	cfg, err := config.Load(configPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: %v (using defaults)\n", err)
	}

	// Register any user-defined custom themes from config
	for _, customDef := range cfg.CustomThemes {
		th := customDef.BuildTheme(theme.Default())
		theme.Register(th)
	}

	// Resolve active theme: CLI flag takes precedence over config
	targetThemeName := cfg.Theme
	if *themeName != "" {
		targetThemeName = *themeName
	}
	selectedTheme, _ := theme.Get(targetThemeName)

	// Apply any global color overrides from config
	if cfg.ColorOverrides != nil {
		selectedTheme = config.ApplyOverrides(selectedTheme, cfg.ColorOverrides)
	}

	// Initialize and run Bubble Tea program with alternate screen buffer
	p := tea.NewProgram(
		ui.InitialModel(client, selectedTheme, cfg, configPath),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}
