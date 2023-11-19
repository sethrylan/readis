package main

import "github.com/charmbracelet/bubbles/key"

type listKeyMap struct {
	ScanMore   key.Binding
	CursorUp   key.Binding
	CursorDown key.Binding
	PageNext   key.Binding
	PagePrev   key.Binding

	// toggleSpinner    key.Binding
	// toggleTitleBar   key.Binding
	// toggleStatusBar  key.Binding
	// togglePagination key.Binding
	// toggleHelpMenu   key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		ScanMore: key.NewBinding(
			key.WithKeys("ctrl+m"),
			key.WithHelp("ctrl+m", "scan more"),
		),
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

		// insertItem: key.NewBinding(
		// 	key.WithKeys("a"),
		// 	key.WithHelp("a", "add item"),
		// ),
		// toggleSpinner: key.NewBinding(
		// 	key.WithKeys("s"),
		// 	key.WithHelp("s", "toggle spinner"),
		// ),
		// toggleTitleBar: key.NewBinding(
		// 	key.WithKeys("T"),
		// 	key.WithHelp("T", "toggle title"),
		// ),
		// toggleStatusBar: key.NewBinding(
		// 	key.WithKeys("S"),
		// 	key.WithHelp("S", "toggle status"),
		// ),
		// togglePagination: key.NewBinding(
		// 	key.WithKeys("P"),
		// 	key.WithHelp("P", "toggle pagination"),
		// ),
		// toggleHelpMenu: key.NewBinding(
		// 	key.WithKeys("H"),
		// 	key.WithHelp("H", "toggle help"),
		// ),
	}
}
