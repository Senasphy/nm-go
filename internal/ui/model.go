package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/Senasphy/nm-go/internal/config"
	"github.com/Senasphy/nm-go/internal/iwd"
	"github.com/Senasphy/nm-go/internal/theme"
)

type TabType int

const (
	TabAvailable TabType = iota
	TabSaved
	TabDetails
)

type NotificationType int

const (
	NotifInfo NotificationType = iota
	NotifSuccess
	NotifWarning
	NotifError
)

type Notification struct {
	Message string
	Type    NotificationType
	Expires time.Time
}

type Model struct {
	client     *iwd.Client
	err        error
	width      int
	height     int

	// Active Theme & Side-by-Side Theme Picker
	currentTheme     theme.Theme
	isThemePicker    bool
	themePickerIndex int

	// Data
	activeTab     TabType
	devices       []iwd.Device
	selectedDev   int
	station       *iwd.Station
	networks      []iwd.Network
	knownNetworks []iwd.KnownNetwork
	activeNetwork *iwd.Network

	// Navigation & Selection
	cursorAvailable int
	cursorSaved     int

	// Filtering
	filterInput textinput.Model
	filtering   bool

	// Modals & Popups (strictly for forms)
	modal         ModalType
	passwordModal PasswordModal
	targetForget  *iwd.KnownNetwork

	// Configuration
	cfg        *config.Config
	configPath string

	// Status & Notifications
	notification *Notification
	spinner      spinner.Model
	isScanning   bool
	isConnecting bool
	busyAction   string
}

// InitialModel initializes the Bubble Tea model with a chosen theme and configuration
func InitialModel(client *iwd.Client, th theme.Theme, cfg *config.Config, configPath string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	fi := textinput.New()
	fi.Placeholder = "Type to filter SSIDs..."
	fi.CharLimit = 32
	fi.Width = 30

	initialTab := TabAvailable
	if cfg != nil {
		switch strings.ToLower(cfg.DefaultTab) {
		case "saved":
			initialTab = TabSaved
		case "diagnostics", "device", "info", "details":
			initialTab = TabDetails
		}
	}

	return Model{
		client:          client,
		currentTheme:    th,
		cfg:             cfg,
		configPath:      configPath,
		activeTab:       initialTab,
		cursorAvailable: 0,
		cursorSaved:     0,
		spinner:         s,
		filterInput:     fi,
		modal:           ModalNone,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchDataCmd(m.client, false),
		startPeriodicTick(),
	)
}

func startPeriodicTick() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return PeriodicTickMsg(t)
	})
}
