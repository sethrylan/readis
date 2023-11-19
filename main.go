package main

import (
	"fmt"
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

type model struct {
	data         *Data
	keyMap       *listKeyMap
	patternInput textinput.Model
	keylist      list.Model
}

func panicOnError[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func initialModel() model {
	m := model{}

	m.data = NewData()
	m.keyMap = newListKeyMap()

	m.patternInput = textinput.New()
	m.patternInput.Cursor.Style = cursorStyle
	m.patternInput.CharLimit = 32
	m.patternInput.Placeholder = "Pattern"
	m.patternInput.Focus()
	m.patternInput.PromptStyle = focusedStyle
	m.patternInput.TextStyle = focusedStyle

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	m.keylist = list.New([]list.Item{}, delegate, 0, 0)
	m.keylist.SetStatusBarItemName("Key", "Keys")
	m.keylist.SetShowTitle(false)
	m.keylist.SetShowPagination(true)
	m.keylist.SetFilteringEnabled(false)
	m.keylist.SetHeight(docStyle.GetHeight() - 10)
	m.keylist.KeyMap.CursorUp = m.keyMap.CursorUp
	m.keylist.KeyMap.CursorDown = m.keyMap.CursorDown
	m.keylist.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			m.keyMap.ScanMore,
			m.keyMap.PageNext,
			m.keyMap.PagePrev,
		}
	}

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
			return m, tea.Quit // TODO close data
		case "enter":
			m.data.ResetScan()
			// m.keysScanned, m.keysTotal, items = m.data.ScanMock(panicOnError(strconv.Atoi(m.patternInput.Value())))
			items := m.data.NewScan(m.patternInput.Value(), 10)
			m.keylist.SetItems(items)
			var cmd tea.Cmd
			m.keylist, cmd = m.keylist.Update(msg)
			return m, tea.Batch(cmd)
		case "up", "down", "left", "right":
			var cmd tea.Cmd
			m.keylist, cmd = m.keylist.Update(msg)
			return m, tea.Batch(cmd)
		case "ctrl+m":
			m.data.ScanMore()
			// m.keylist.SetShowHelp(!m.keylist.ShowHelp())
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		patternInputHeight := 4 // TODO: calculate this value from height + margin
		m.keylist.SetSize(msg.Width-h, msg.Height-v-patternInputHeight)
	}

	// Handle character input and blinking
	var cmd tea.Cmd
	m.patternInput, cmd = m.patternInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString(
		lipgloss.NewStyle().MarginBottom(2).Render(
			lipgloss.JoinHorizontal(
				lipgloss.Center,
				lipgloss.NewStyle().Height(2).Width(30).Render(m.patternInput.View()),
				lipgloss.NewStyle().PaddingLeft(10).Render(
					lipgloss.JoinVertical(lipgloss.Right,
						lipgloss.NewStyle().Render(m.data.opts.Addrs[0]),
						fmt.Sprintf("%d keys", m.data.TotalKeys()),
					),
				),
			),
		),
	)
	// b.WriteString(helpStyle.Render(fmt.Sprintf("%d Matches", m.data.TotalFound())))

	return lipgloss.JoinVertical(lipgloss.Left,
		b.String(),
		docStyle.Render(m.keylist.View()))
}

////////////////////////////////////

type Key struct {
	name    string
	keyType string // Hash, String, Set, etc.
	size    int64  // in bytes
	ttl     time.Duration
}

func (k Key) Title() string {
	return lipgloss.NewStyle().Width(11).Render(lipgloss.NewStyle().Background(ColorForKeyType(k.keyType)).Render(k.keyType)) +
		lipgloss.NewStyle().Width(25).Render(k.name) +
		lipgloss.NewStyle().Width(12).Render(strconv.FormatInt(k.size, 10)+"B") +
		lipgloss.NewStyle().Width(8).Render(fmt.Sprintf("%.0fs", k.ttl.Seconds()))
}

func (k Key) Description() string {
	return ""
}

func (k Key) FilterValue() string {
	return k.name
}

////////////////////////////////////////////

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}
