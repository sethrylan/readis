// Package ui provides user interface components.
package ui

import "charm.land/bubbles/v2/key"

// ListKeyMap defines key bindings for the list component.
type ListKeyMap struct {
	CursorUp   key.Binding
	CursorDown key.Binding
	PageNext   key.Binding
	PagePrev   key.Binding
	GoToStart  key.Binding
	GoToEnd    key.Binding
}

// NewListKeyMap creates a new ListKeyMap with adjusted keys for fewer letter keys.
func NewListKeyMap() *ListKeyMap {
	return &ListKeyMap{
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
