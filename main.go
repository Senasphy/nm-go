package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"nm-go/internal/iwd"
	"nm-go/internal/ui"
)

const version = "0.1.0"

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	showHelp := flag.Bool("help", false, "Print help and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("nm-go v%s (Wi-Fi TUI for iwd)\n", version)
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

	// Initialize and run Bubble Tea program with alternate screen buffer
	p := tea.NewProgram(
		ui.InitialModel(client),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}
