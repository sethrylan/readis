package main

import "github.com/charmbracelet/bubbles/key"

type listKeyMap struct {
	CursorUp   key.Binding
	CursorDown key.Binding
	PageNext   key.Binding
	PagePrev   key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		PageNext: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "next page"),
		),
		PagePrev: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "previous page"),
		),
		CursorUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
	}
}
