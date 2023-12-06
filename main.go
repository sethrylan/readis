package main

import (
	"context"
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
	initialized  bool
	scan         *Scan
	scanCh       chan *Key
	scanCtx      context.Context
}

var logfile *os.File

type (
	errMsg error
)

func debug(a ...string) {
	if logfile == nil {
		return
	}

	for _, s := range a {
		logfile.WriteString(s)
	}

	logfile.WriteString("\n")
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
	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		debug("key pressed: ", msg.String())
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.data.Close()
			return m, tea.Quit
		case "enter":
			// Initialize scan, estimate number of keys per page
			m.scan = m.data.NewScan(m.patternInput.Value(), m.keylist.Paginator.ItemsOnPage(1000))
			m.scanCh = make(chan *Key)
			m.scanCtx = context.Background()
			var cmds []tea.Cmd
			m.data.asyncScan(m.scanCtx, m.scan, m.scanCh)
			for key := range m.scanCh {
				c := m.keylist.InsertItem(10000000000, *key)
				cmds = append(cmds, c)
			}

			var cmd tea.Cmd
			m.keylist, cmd = m.keylist.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case "up", "down", "left":
			var cmd tea.Cmd
			m.keylist, cmd = m.keylist.Update(msg)
			setViewportContent(&m)
			return m, tea.Batch(cmd)
		case "ctrl+t", "right":
			var cmds []tea.Cmd
			if m.keylist.Paginator.OnLastPage() {
				m.scanCh = make(chan *Key)
				m.scanCtx = context.Background()

				m.data.asyncScan(m.scanCtx, m.scan, m.scanCh)
				for key := range m.scanCh {
					c := m.keylist.InsertItem(10000000000, *key)
					cmds = append(cmds, c)
				}
			}
			var cmd tea.Cmd
			m.keylist, cmd = m.keylist.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}
	case tea.WindowSizeMsg:
		// WindowSizeMsg is sent before the first render and then again every resize.

		horizontalMargin, verticalMargin := docStyle.GetFrameSize() // horizontal and vertical margins
		keylistWidth := msg.Width - horizontalMargin
		keylistHeight := msg.Height - verticalMargin - lipgloss.Height(m.headerView())
		m.keylist.SetSize(keylistWidth, keylistHeight)
		headerHeight := lipgloss.Height(m.headerView())

		viewportWidth := msg.Width - horizontalMargin - 112 // the sum of Title widths and spacing (or input style width)
		viewportHeight := keylistHeight - 5                 // adjust for spacing
		m.viewport = viewport.New(viewportWidth, viewportHeight)
		m.viewport.Style = viewportStyle.Width(viewportWidth)
		m.viewport.YPosition = headerHeight
		setViewportContent(&m)
		statusBlockStyle = statusBlockStyle.Width(viewportWidth)

		m.initialized = true
	case errMsg:
		// handle errors like any other message
		m.viewport.SetContent(msg.Error())
		return m, nil
	}

	// Handle any other character input as pattern input
	var cmd tea.Cmd
	m.patternInput, cmd = m.patternInput.Update(msg)
	return m, cmd
}

func (m model) headerView() string {
	input := inputStyle.Render(m.patternInput.View())
	statusBlock := statusBlockStyle.Render(
		lipgloss.JoinVertical(lipgloss.Right,
			m.data.opts.Addrs[0],
			fmt.Sprintf("%d keys", m.data.TotalKeys(context.Background())),
		),
	)

	return lipgloss.NewStyle().Render(
		lipgloss.JoinHorizontal(lipgloss.Top, input, statusBlock),
	)
}

func (m model) resultsView() string {
	if m.keylist.SelectedItem() == nil {
		return m.keylist.View()
	}

	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.keylist.View(),
		m.viewport.View(),
	)
}

func setViewportContent(m *model) {
	if m.keylist.SelectedItem() != nil {
		markdown := m.data.Fetch(m.keylist.SelectedItem().(Key))
		renderer := panicOnError(glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.viewport.Width),
		))

		str := panicOnError(renderer.Render(markdown))
		m.viewport.SetContent(str)
	}
}

func (m model) View() string {
	if !m.initialized {
		return "\n  Initializing..."
	}

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

func (k Key) TTLString() string {
	if k.ttl == -1 {
		return "∞"
	}
	return humanize.RelTime(time.Now(), time.Now().Add(k.ttl), "", "")
}

func (k Key) Title() string {
	return lipgloss.NewStyle().Width(10).Render(lipgloss.NewStyle().Background(ColorForKeyType(k.datatype)).Render(k.datatype)) +
		lipgloss.NewStyle().Width(80).Render(k.name) +
		lipgloss.NewStyle().Width(11).Render(k.TTLString()) +
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
	if len(os.Getenv("DEBUG")) > 0 {
		logfile = panicOnError(tea.LogToFile("debug.log", "debug"))
		defer logfile.Close()
	}

	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(), // use the full size of the terminal in the alternate screen buffer
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}
