package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/sethrylan/readis/internal/data"
	"github.com/sethrylan/readis/internal/ui"
	"github.com/sethrylan/readis/internal/util"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// appCtx and appCancel manage the application lifecycle context.
// They live at package level rather than in the model struct to
// satisfy the containedctx linter while remaining accessible to
// Bubble Tea callbacks that do not receive a context parameter.
var appCtx, appCancel = context.WithCancel(context.Background()) //nolint:gochecknoglobals

type model struct {
	data    *data.Data
	scan    *data.Scan
	scanCh  <-chan *data.Key // receive-only channel for scan results
	spinner spinner.Model

	textinput   textinput.Model
	keylist     list.Model
	viewport    viewport.Model
	initialized bool
	totalKeys   int64

	windowHeight, windowWidth int
}

type totalKeysMsg int64

type refreshTotalKeysMsg struct{}

func tickTotalKeys() tea.Cmd {
	return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
		return refreshTotalKeysMsg{}
	})
}

// resizeViews should be called anytime the panes need to be updated.
// We can think of the rendered UI as having four panes:
//
// |---------------------|
// |  input  |  status   |
// | ------------------- |
// | keylist | viewport  |
// |---------------------|
//
// We call the top too the "header". The separations between the panes
// should be kept consistent. That means we need to resize when:
// 1) the window is reszied
// 2) the keylist is updated (because the longer key names may require more on the left hand side)
// 3) we change pages (same reason as 2)
//
// So we keep track of the longest key name and the window size for resizing.
func (m *model) resizeViews() {
	// Find the longest key name, we'll use that to resize the left hand pane
	for _, k := range m.keylist.VisibleItems() {
		if k, ok := k.(keyItem); ok {
			keyNameWidth = max(keyNameWidth, len(k.Name)+1)
		}
	}

	hMargin, vMargin := docStyle.GetFrameSize()
	headerHeight := lipgloss.Height(m.headerView())
	keylistWidth := leftHandWidth()
	keylistHeight := m.windowHeight - vMargin - headerHeight
	m.keylist.SetSize(keylistWidth, keylistHeight)

	util.Debug(fmt.Sprintf("keyNameWidth: %d", keyNameWidth))
	util.Debug(fmt.Sprintf("window width: %d, height: %d", m.windowWidth, m.windowHeight))
	util.Debug(fmt.Sprintf("frame width: %d, height: %d", hMargin, vMargin))
	util.Debug(fmt.Sprintf("keylist width: %d, height: %d", keylistWidth, keylistHeight))

	// Update rightHandWidth (also used for styling the status block)
	rightHandWidth = m.windowWidth - hMargin - leftHandWidth()

	viewportWidth := rightHandWidth
	viewportHeight := keylistHeight - headerHeight
	m.viewport = viewport.New(viewportWidth, viewportHeight)
	m.viewport.Style = viewportStyle.Width(viewportWidth)
	m.viewport.YPosition = headerHeight
	m.setViewportContent(appCtx)
}

func newModel(d *data.Data) *model {
	m := &model{}

	km := ui.NewListKeyMap()
	m.data = d

	m.spinner = spinner.New(
		spinner.WithSpinner(spinner.Ellipsis),
		spinner.WithStyle(spinnerStyle),
	)

	m.textinput = textinput.New()
	m.textinput.Cursor.Style = cursorStyle
	m.textinput.CharLimit = 80
	m.textinput.Placeholder = "Pattern"
	m.textinput.Focus()
	m.textinput.PromptStyle = focusedStyle
	m.textinput.TextStyle = focusedStyle

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	m.keylist = list.New([]list.Item{}, delegate, 0, 0)
	m.keylist.SetStatusBarItemName("Key", "Keys")
	m.keylist.SetShowStatusBar(false)
	m.keylist.SetShowTitle(false)
	m.keylist.Help.ShowAll = false
	m.keylist.SetShowPagination(true)
	m.keylist.SetFilteringEnabled(false)
	m.keylist.KeyMap.CursorUp = km.CursorUp
	m.keylist.KeyMap.CursorDown = km.CursorDown
	m.keylist.KeyMap.GoToStart = km.GoToStart
	m.keylist.KeyMap.GoToEnd = km.GoToEnd
	m.keylist.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			km.PageNext,
			km.PagePrev,
		}
	}
	m.keylist.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			km.PageNext,
			km.PagePrev,
		}
	}
	return m
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick, m.refreshTotalKeys, tickTotalKeys())
}

func (m *model) refreshTotalKeys() tea.Msg {
	return totalKeysMsg(m.data.TotalKeys(appCtx))
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0, 4)
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case totalKeysMsg:
		m.totalKeys = int64(msg)
		return m, nil
	case refreshTotalKeysMsg:
		return m, tea.Batch(m.refreshTotalKeys, tickTotalKeys())
	case tea.KeyMsg:
		util.Debug("key pressed: ", msg.String())
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			appCancel()
			err := m.data.Close()
			if err != nil {
				fmt.Println("error closing connection: ", err)
			}
			return m, tea.Quit
		case "enter":
			m.keylist.SetItems([]list.Item{})                    // clear items
			pageSize := m.keylist.Paginator.ItemsOnPage(1000)    // estimate the page size
			m.scan = data.NewScan(m.textinput.Value(), pageSize) // initialize scan
			m.scanCh = m.data.ScanAsync(appCtx, m.scan)          // start scan
			m.keylist, cmd = m.keylist.Update(msg)
			return m, tea.Batch(append(cmds, cmd)...)
		case "up", "down", "left", "?", "home", "end", "pgdown", "pgup":
			m.keylist, cmd = m.keylist.Update(msg)
			m.resizeViews()
			return m, tea.Batch(cmd)
		case "ctrl+t", "right":
			// If on the last page and the current scan is complete,
			// then we can scan for the next page of results.
			// And ctrl+t? That's just an undocumented shortcut.
			if m.keylist.Paginator.OnLastPage() && m.scan != nil && !m.scan.Scanning() && m.scan.HasMore() {
				m.scanCh = m.data.ScanAsync(appCtx, m.scan)
			}
			m.keylist, cmd = m.keylist.Update(msg)
			m.resizeViews()
			return m, tea.Batch(append(cmds, cmd)...)
		}
	case tea.WindowSizeMsg:
		// WindowSizeMsg is sent before the first render and then again every resize.
		m.windowHeight, m.windowWidth = msg.Height, msg.Width
		m.resizeViews()
		m.initialized = true
	case error:
		// handle errors like any other message
		m.viewport.SetContent(msg.Error())
		return m, nil
	}

	cmds = append(cmds, m.readAndInsert()...)

	if m.viewport.VisibleLineCount() == 0 {
		// On new searches, update the viewport with the first list item.
		m.setViewportContent(appCtx)
	}

	// Handle any other character input as pattern input
	m.textinput, cmd = m.textinput.Update(msg)
	cmds = append(cmds, cmd)

	// Tick the spinner
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) readAndInsert() []tea.Cmd {
	var cmds []tea.Cmd
	for {
		select {
		case k, ok := <-m.scanCh:
			if !ok {
				return cmds
			}
			util.Debug("found key: ", k.Name)
			cmd := m.keylist.InsertItem(math.MaxInt, keyItem{*k})
			cmds = append(cmds, cmd)
		default:
			return cmds
		}
	}
}

func (m *model) spinnerView() string {
	if m.scan == nil || !m.scan.Scanning() {
		return " "
	}
	return spinnerStyle.Render("   scanning") + m.spinner.View()
}

func (m *model) headerView() string {
	inputBlock := headerStyle.
		Width(leftHandWidth() - 6).
		Align(lipgloss.Left).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			m.textinput.View(),
			m.spinnerView(),
		))
	statusBlock := headerStyle.
		Width(rightHandWidth).
		Align(lipgloss.Right).
		Render(lipgloss.JoinVertical(lipgloss.Right,
			m.data.URI(),
			fmt.Sprintf("%d keys", m.totalKeys),
		))

	return lipgloss.NewStyle().Render(
		lipgloss.JoinHorizontal(lipgloss.Top, inputBlock, statusBlock),
	)
}

func (m *model) resultsView() string {
	if m.keylist.SelectedItem() == nil {
		return m.keylist.View()
	}

	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.keylist.View(),
		m.viewport.View(),
	)
}

func (m *model) setViewportContent(ctx context.Context) {
	if m.keylist.SelectedItem() != nil {
		selectedKey, ok := m.keylist.SelectedItem().(keyItem)
		if !ok {
			return
		}
		markdown, err := m.data.Fetch(ctx, selectedKey.Key)
		if err != nil {
			m.viewport.SetContent(err.Error())
			return
		}
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.viewport.Width),
		)
		if err != nil {
			m.viewport.SetContent(err.Error())
			return
		}

		str, err := renderer.Render(markdown)
		if err != nil {
			m.viewport.SetContent(err.Error())
			return
		}
		m.viewport.SetContent(str)
	}
}

func (m *model) View() string {
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
