package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Focus Areas
const (
	PatternInput int = iota
	KeyList
)

func randtype() string {
	types := []string{
		"set",
		"sorted set",
		"hash",
		"string",
		"list",
	}
	n := rand.Int() % len(types)
	return types[n]

}

var allkeys = [...]list.Item{
	rkey{key: "Raspberry Pi’s", keyType: randtype(), size: rand.Intn(100), ttl: time.Duration(rand.Intn(100000000000))},
	rkey{key: "Nutella", keyType: randtype(), size: 12, ttl: 0},
	rkey{key: "Bitter melon", keyType: randtype(), size: 12, ttl: 0},
	rkey{key: "Nice socks", keyType: randtype(), size: 12, ttl: 0},
	rkey{key: "Eight hours of sleep", keyType: randtype(), size: 12, ttl: 0},
	rkey{key: "Cats", keyType: randtype(), size: 12, ttl: 0},
	rkey{key: "Plantasia, the album", keyType: randtype(), size: 12, ttl: 0},
	rkey{key: "Pour over coffee", keyType: randtype(), size: 12, ttl: 0},
	rkey{key: "VR", keyType: randtype(), size: 12, ttl: 0},
	rkey{key: "Noguchi Lamps", keyType: "hash", size: 12, ttl: 0},
	rkey{key: "Linux", keyType: "hash", size: 12, ttl: 0},
	rkey{key: "Business school", keyType: "hash", size: 12, ttl: 0},
	rkey{key: "Pottery", keyType: "hash", size: 12, ttl: 0},
	rkey{key: "Shampoo", keyType: "hash", size: 12, ttl: 0},
	rkey{key: "Table tennis", keyType: "hash", size: 12, ttl: 0},
	rkey{key: "Milk crates", keyType: "hash", size: 12, ttl: 0},
	rkey{key: "Afternoon tea", keyType: "hash", size: 12, ttl: 0},
	rkey{key: "Stickers", keyType: "hash", size: 12, ttl: 0},
	rkey{key: "20° Weather", keyType: "hash", size: 12, ttl: 0},
	rkey{key: "Warm light", keyType: "hash", size: 12, ttl: 0},
	rkey{key: "The vernal equinox", keyType: "hash", size: 12, ttl: 0},
	rkey{key: "Gaffer’s tape", keyType: "hash", size: 12, ttl: 0},
	rkey{key: "Terrycloth", keyType: "hash", size: 12, ttl: 0},
}

type model struct {
	focus  int
	keyMap *listKeyMap

	patternInput textinput.Model

	rkeyList list.Model
}

func initialModel() model {
	m := model{}

	m.keyMap = newListKeyMap()

	m.patternInput = textinput.New()
	m.patternInput.Cursor.Style = cursorStyle
	m.patternInput.CharLimit = 32
	m.patternInput.Placeholder = "Pattern"
	m.patternInput.Focus()
	m.patternInput.PromptStyle = focusedStyle
	m.patternInput.TextStyle = focusedStyle

	m.rkeyList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.rkeyList.SetStatusBarItemName("Key", "Keys")
	m.rkeyList.SetShowTitle(false)
	m.rkeyList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			m.keyMap.scanMore,
		}
	}

	m.rkeyList.SetHeight(docStyle.GetHeight() - 10)

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		case "enter":
			// Did the user press enter while pattern input was focused?
			// TODO: run search
			if m.patternInput.Focused() {
				i, err := strconv.Atoi(m.patternInput.Value())
				i = min(i, len(allkeys))
				if err == nil {
					m.rkeyList.SetItems(allkeys[:i])
				}
			}
			var cmd tea.Cmd
			m.rkeyList, cmd = m.rkeyList.Update(msg)
			return m, tea.Batch(cmd)

		case "up", "down":
			var cmd tea.Cmd
			m.rkeyList, cmd = m.rkeyList.Update(msg)
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
		patternInputHeight := 3 // TODO: calculate this value
		m.rkeyList.SetSize(msg.Width-h, msg.Height-v-patternInputHeight)
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
	fmt.Fprintf(&b, "\n")

	b.WriteString(helpStyle.Render("Scanned", "123", "of 412345"))

	return lipgloss.JoinVertical(lipgloss.Left,
		b.String(),
		docStyle.Render(m.rkeyList.View()))
}

////////////////////////////////////

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type rkey struct {
	key     string
	keyType string // Hash, String, Set, etc.
	size    int    // in bytes
	ttl     time.Duration
}

func (i rkey) Title() string {
	return i.key
}

func (i rkey) Description() string {
	return lipgloss.NewStyle().Background(ColorForKeyType(i.keyType)).Render(i.keyType) +
		" " +
		lipgloss.NewStyle().Width(20).Render(i.key) +
		" " +
		lipgloss.NewStyle().Width(20).Render(strconv.Itoa(i.size)+"bytes") +
		fmt.Sprintf("%.0f", i.ttl.Seconds()) + "s"
}

func (i rkey) FilterValue() string {
	return i.key
}

////////////////////////////////////////////

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}
