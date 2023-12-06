package main

import "github.com/charmbracelet/lipgloss"

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#c9510c"))
	// blurredStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#c9510c"))
	cursorStyle = focusedStyle.Copy()
	// helpStyle        = blurredStyle.Copy()
	docStyle         = lipgloss.NewStyle().Margin(1, 2)
	headerStyle      = lipgloss.NewStyle().Margin(1, 1).Foreground(lipgloss.Color("#c9510c")).Bold(true).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#0a2b3b"))
	inputStyle       = headerStyle.Copy().Width(104)
	statusBlockStyle = headerStyle.Copy().Align(lipgloss.Right)

	viewportStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6e5494")).
			PaddingRight(2)
	// noStyle      = lipgloss.NewStyle()
	// cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	// focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	// blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

func ColorForKeyType(keyType string) lipgloss.Color {
	switch keyType {
	case "hash":
		return lipgloss.Color("#0000ff")
	case "set":
		return lipgloss.Color("#935f35")
	case "sorted set":
		return lipgloss.Color("#932069")
	case "string":
		return lipgloss.Color("#6123bc")
	default:
		return lipgloss.Color("#00ff00")
	}
}
