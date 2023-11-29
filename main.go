package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

type model struct {
	data         *Data
	keyMap       *listKeyMap
	patternInput textinput.Model
	keylist      list.Model
	viewport     viewport.Model
}

// TODO: errMsg https://github.com/charmbracelet/bubbletea/blob/a6f07b8ba6439fa65612a350bc1878d9d8c0447a/examples/chat/main.go#L26

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
	m.patternInput.CharLimit = 77
	m.patternInput.Placeholder = "Pattern"
	m.patternInput.Focus()
	m.patternInput.PromptStyle = focusedStyle
	m.patternInput.TextStyle = focusedStyle

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	m.keylist = list.New([]list.Item{}, delegate, 0, 0)
	m.keylist.SetStatusBarItemName("Key", "Keys")
	m.keylist.SetShowStatusBar(false)
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

	m.viewport = newvalueview()
	// m.viewport.HighPerformanceRendering = true // TODO
	m.viewport.Height = m.keylist.Height()

	return m
}

func newvalueview() viewport.Model {
	vp := viewport.New(80, 30)
	vp.Style = viewportStyle
	return vp
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, viewport.Sync(m.viewport))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.data.Close()
			return m, tea.Quit
		case "enter":
			m.data.ResetScan()
			items := m.data.NewScan(m.patternInput.Value(), 10)
			m.keylist.SetItems(items)
			var cmd tea.Cmd
			m.keylist, cmd = m.keylist.Update(msg)
			return m, tea.Batch(cmd)
		case "up", "down", "left", "right":
			var cmd tea.Cmd
			m.keylist, cmd = m.keylist.Update(msg)

			if m.keylist.SelectedItem() != nil {
				markdown := m.data.Fetch(m.keylist.SelectedItem().(Key))
				renderer := panicOnError(glamour.NewTermRenderer(
					glamour.WithAutoStyle(),
					glamour.WithWordWrap(m.viewport.Width),
				))

				str := panicOnError(renderer.Render(markdown))

				m.viewport.SetContent(str)

			}

			return m, tea.Batch(cmd)
		case "ctrl+m":
			m.data.ScanMore()
			// m.keylist.SetShowHelp(!m.keylist.ShowHelp())
		}
	case tea.WindowSizeMsg:
		// Note that WindowSizeMsg is sent before the first render and then again every resize.

		h, v := docStyle.GetFrameSize()
		patternInputHeight := headerStyle.GetVerticalFrameSize()
		m.keylist.SetSize(msg.Width-h, msg.Height-v-patternInputHeight)
		m.viewport.Height = m.keylist.Height() - 5 // adjust for pagination and help message
		// TODO: dynamic viewport resizing: https://github.com/charmbracelet/bubbletea/blob/a6f07b8ba6439fa65612a350bc1878d9d8c0447a/examples/pager/main.go#L71-L75
	}

	// Handle character input
	var cmd tea.Cmd
	m.patternInput, cmd = m.patternInput.Update(msg)
	return m, cmd
}

func (m model) headerView() string {
	input := headerStyle.Copy().Width(105).Render(m.patternInput.View())
	statusBlock := statusBlockStyle.Render(
		lipgloss.JoinVertical(lipgloss.Right,
			m.data.opts.Addrs[0],
			fmt.Sprintf("%d keys", m.data.TotalKeys()),
		),
	)

	return lipgloss.NewStyle().Render(
		lipgloss.JoinHorizontal(lipgloss.Top, input, statusBlock),
	)
}

func (m model) resultsView() string {
	if len(m.keylist.VisibleItems()) == 0 {
		return m.keylist.View()
	}

	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.keylist.View(),
		m.viewport.View(),
	)
}

func (m model) View() string {

	// b.WriteString(helpStyle.Render(fmt.Sprintf("%d Matches", m.data.TotalFound())))

	return docStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			m.headerView(),
			m.resultsView(),
		),
	)

}

////////////////////////////////////

type Key struct {
	name     string
	datatype string // Hash, String, Set, etc; https://redis.io/commands/type/
	size     uint64 // in bytes
	ttl      time.Duration
}

func (k Key) Title() string {
	var ttl string
	if k.ttl == -1 {
		ttl = "∞"
	} else {
		ttl = humanize.RelTime(time.Now(), time.Now().Add(k.ttl), "", "")
	}
	return lipgloss.NewStyle().Width(10).Render(lipgloss.NewStyle().Background(ColorForKeyType(k.datatype)).Render(k.datatype)) +
		lipgloss.NewStyle().Width(80).Render(k.name) +
		lipgloss.NewStyle().Width(11).Render(ttl) +
		lipgloss.NewStyle().Width(7).Render(humanize.Bytes(k.size))
}

func (k Key) Description() string {
	return ""
}

func (k Key) FilterValue() string {
	return k.name
}

////////////////////////////////////////////

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(), // use the full size of the terminal in its "alternate screen buffer"
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}
