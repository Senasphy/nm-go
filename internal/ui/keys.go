package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Tab        key.Binding
	BackTab    key.Binding
	Connect    key.Binding
	Disconnect key.Binding
	Scan       key.Binding
	Forget     key.Binding
	Power      key.Binding
	AutoConn   key.Binding
	Filter     key.Binding
	Help       key.Binding
	Quit       key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	BackTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev tab"),
	),
	Connect: key.NewBinding(
		key.WithKeys("enter", "c"),
		key.WithHelp("enter/c", "connect"),
	),
	Disconnect: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "disconnect"),
	),
	Scan: key.NewBinding(
		key.WithKeys("s", "r"),
		key.WithHelp("s/r", "scan/refresh"),
	),
	Forget: key.NewBinding(
		key.WithKeys("f", "x"),
		key.WithHelp("f", "forget network"),
	),
	Power: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "toggle power"),
	),
	AutoConn: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "toggle autoconnect"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
