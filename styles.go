package main

import "github.com/charmbracelet/lipgloss"

var (
	TypeLabelWidth = 10 // max is "string"
	KeyNameWidth   = 20 // assume the max to start, and adjust as keys are found
	TTLWidth       = 12 // max is "101 minutes"
	SizeWidth      = 7
	RightHandWidth = 30
)

func LeftHandWidth() int {
	return TypeLabelWidth + KeyNameWidth + TTLWidth + SizeWidth + 3
}

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#c9510c"))
	cursorStyle  = focusedStyle.Copy()
	docStyle     = lipgloss.NewStyle().Margin(1, 2)
	headerStyle  = lipgloss.NewStyle().
			Margin(0, 1, 1).
			Foreground(lipgloss.Color("#c9510c")).
			Bold(true).
			Border(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("#0a2b3b"))
	viewportStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("#6e5494")).
			PaddingRight(2)
	spinnerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00ff00"))
)

func ColorForKeyType(keyType string) lipgloss.Color {
	switch keyType {
	case "hash":
		return lipgloss.Color("#0000ff")
	case "set":
		return lipgloss.Color("#935f35")
	case "zset":
		return lipgloss.Color("#932069")
	case "string":
		return lipgloss.Color("#6123bc")
	default:
		return lipgloss.Color("#00ff00")
	}
}
