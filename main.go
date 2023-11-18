package main

// A simple example demonstrating the use of multiple text input components
// from the Bubbles component library.

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	// focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	// blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

// Focus Areas
const (
	PatternInput int = iota
	KeyList
)

var allkeys = [...]list.Item{
	key{title: "Raspberry Pi’s", desc: "I have ’em all over my house"},
	key{title: "Nutella", desc: "It's good on toast"},
	key{title: "Bitter melon", desc: "It cools you down"},
	key{title: "Nice socks", desc: "And by that I mean socks without holes"},
	key{title: "Eight hours of sleep", desc: "I had this once"},
	key{title: "Cats", desc: "Usually"},
	key{title: "Plantasia, the album", desc: "My plants love it too"},
	key{title: "Pour over coffee", desc: "It takes forever to make though"},
	key{title: "VR", desc: "Virtual reality...what is there to say?"},
	key{title: "Noguchi Lamps", desc: "Such pleasing organic forms"},
	key{title: "Linux", desc: "Pretty much the best OS"},
	key{title: "Business school", desc: "Just kidding"},
	key{title: "Pottery", desc: "Wet clay is a great feeling"},
	key{title: "Shampoo", desc: "Nothing like clean hair"},
	key{title: "Table tennis", desc: "It’s surprisingly exhausting"},
	key{title: "Milk crates", desc: "Great for packing in your extra stuff"},
	key{title: "Afternoon tea", desc: "Especially the tea sandwich part"},
	key{title: "Stickers", desc: "The thicker the vinyl the better"},
	key{title: "20° Weather", desc: "Celsius, not Fahrenheit"},
	key{title: "Warm light", desc: "Like around 2700 Kelvin"},
	key{title: "The vernal equinox", desc: "The autumnal equinox is pretty good too"},
	key{title: "Gaffer’s tape", desc: "Basically sticky fabric"},
	key{title: "Terrycloth", desc: "In other words, towel fabric"},
}

type model struct {
	focus int

	patternInput textinput.Model
	cursorMode   cursor.Mode

	keyList list.Model
}

func initialModel() model {
	m := model{}

	m.patternInput = textinput.New()
	m.patternInput.Cursor.Style = cursorStyle
	m.patternInput.CharLimit = 32
	m.patternInput.Placeholder = "Pattern"
	m.patternInput.Focus()
	m.patternInput.PromptStyle = focusedStyle
	m.patternInput.TextStyle = focusedStyle

	m.keyList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.keyList.Title = "Results"

	m.keyList.SetHeight(docStyle.GetHeight() - 10)

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		// Change cursor mode
		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			patternInputCmd := m.patternInput.Cursor.SetMode(m.cursorMode)
			return m, tea.Batch(patternInputCmd)

		case "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while pattern input was focused?
			// TODO: run search
			if s == "enter" && m.patternInput.Focused() {
				i, err := strconv.Atoi(m.patternInput.Value())
				i = min(i, len(allkeys))
				if err == nil {
					m.keyList.SetItems(allkeys[:i])
				}
			}
			var cmd tea.Cmd
			m.keyList, cmd = m.keyList.Update(msg)
			return m, tea.Batch(cmd)

		// Set focus to next input
		case "tab", "shift+tab":
			s := msg.String()

			// Cycle focus
			if s == "shift+tab" {
				m.focus--
			} else {
				m.focus++
			}

			if m.focus > 1 {
				m.focus = 0
			} else if m.focus < 0 {
				m.focus = 1
			}
			var cmd tea.Cmd

			// Set focus to the input at the new focus index
			switch m.focus {
			case PatternInput:
				m.patternInput.Focus()
				m.patternInput.PromptStyle = focusedStyle
				m.patternInput.TextStyle = focusedStyle
			case KeyList:
				m.patternInput.Blur()
				m.patternInput.PromptStyle = noStyle
				m.patternInput.TextStyle = noStyle
			}

			return m, tea.Batch(cmd)

		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		patternInputHeight := 4 // TODO: calculate this value
		m.keyList.SetSize(msg.Width-h, msg.Height-v-patternInputHeight)
	}

	// Handle character input and blinking
	cmd := m.updatePatternInput(msg)

	return m, cmd
}

func (m *model) updatePatternInput(msg tea.Msg) tea.Cmd {
	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.

	var cmd tea.Cmd
	m.patternInput, _ = m.patternInput.Update(msg)
	return tea.Batch(cmd)
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(m.patternInput.View())
	b.WriteRune('\n')

	// button := &blurredButton
	// if m.focusIndex == len(m.inputs) {
	// 	button = &focusedButton
	// }
	fmt.Fprintf(&b, "\n\n")

	// b.WriteString(helpStyle.Render("cursor mode is "))
	// b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	// b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return lipgloss.JoinVertical(lipgloss.Left,
		b.String(),
		docStyle.Render(m.keyList.View()))
}

////////////////////////////////////

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type key struct {
	title, desc string
}

func (i key) Title() string       { return i.title }
func (i key) Description() string { return i.desc }
func (i key) FilterValue() string { return i.title }

////////////////////////////////////////////

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}
