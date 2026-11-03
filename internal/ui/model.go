package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"nm-go/internal/iwd"
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

	// Data
	activeTab      TabType
	devices        []iwd.Device
	selectedDev    int
	station        *iwd.Station
	networks       []iwd.Network
	knownNetworks  []iwd.KnownNetwork
	activeNetwork  *iwd.Network

	// Navigation & Selection
	cursorAvailable int
	cursorSaved     int

	// Filtering
	filterInput textinput.Model
	filtering   bool

	// Modals & Popups
	modal         ModalType
	passwordModal PasswordModal
	targetForget  *iwd.KnownNetwork

	// Status & Notifications
	notification *Notification
	spinner      spinner.Model
	isScanning   bool
	isConnecting bool
	busyAction   string
}

// InitialModel initializes the Bubble Tea model
func InitialModel(client *iwd.Client) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = StyleBadgeMuted

	fi := textinput.New()
	fi.Placeholder = "Type to filter SSIDs..."
	fi.CharLimit = 32
	fi.Width = 24

	return Model{
		client:          client,
		activeTab:       TabAvailable,
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
