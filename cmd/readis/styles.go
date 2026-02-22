package main

import "github.com/charmbracelet/lipgloss"

var (
	typeLabelWidth = 10 // max is "string"
	keyNameWidth   = 20 // assume the max to start, and adjust as keys are found
	ttlWidth       = 12 // max is "101 minutes"
	sizeWidth      = 7
	rightHandWidth = 30
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#c9510c"))
	cursorStyle  = focusedStyle
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

func leftHandWidth() int {
	return typeLabelWidth + keyNameWidth + ttlWidth + sizeWidth + 3
}

func colorForKeyType(keyType string) lipgloss.Color {
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
