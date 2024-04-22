package ui

import "github.com/charmbracelet/bubbles/key"

type listKeyMap struct {
	CursorUp   key.Binding
	CursorDown key.Binding
	PageNext   key.Binding
	PagePrev   key.Binding
	GoToStart  key.Binding
	GoToEnd    key.Binding
}

// These keys adjusted from keys.go, to account for fewer letter keys
func NewListKeyMap() *listKeyMap {
	return &listKeyMap{
		PageNext: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "next"),
		),
		PagePrev: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "prev"),
		),
		CursorUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		GoToStart: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "go to start"),
		),
		GoToEnd: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "go to end"),
		),
	}
}
