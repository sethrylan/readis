package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/github/readis/internal/data"
	"github.com/github/readis/internal/ui"
	"github.com/github/readis/internal/util"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

// ldflags added by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type model struct {
	data    *data.Data
	scan    *data.Scan
	scanCh  <-chan *data.Key // receive-only channel for scan results
	spinner spinner.Model

	textinput   textinput.Model
	keylist     list.Model
	viewport    viewport.Model
	initialized bool

	windowHeight, windowWidth int
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
func (m *model) resizeViews(ctx context.Context) {
	// Find the longest key name, we'll use that to resize the left hand pane
	for _, k := range m.keylist.VisibleItems() {
		if k, ok := k.(Key); ok {
			KeyNameWidth = max(KeyNameWidth, len(k.Name)+1)
		}
	}

	hMargin, vMargin := docStyle.GetFrameSize()
	headerHeight := lipgloss.Height(m.headerView())
	keylistWidth := LeftHandWidth()
	keylistHeight := m.windowHeight - vMargin - headerHeight
	m.keylist.SetSize(keylistWidth, keylistHeight)

	util.Debug(fmt.Sprintf("KeyNameWidth: %d", KeyNameWidth))
	util.Debug(fmt.Sprintf("window width: %d, height: %d", m.windowWidth, m.windowHeight))
	util.Debug(fmt.Sprintf("frame width: %d, height: %d", hMargin, vMargin))
	util.Debug(fmt.Sprintf("keylist width: %d, height: %d", keylistWidth, keylistHeight))

	// Update RightHandWidth (also used for styling the status block)
	RightHandWidth = m.windowWidth - hMargin - LeftHandWidth()

	viewportWidth := RightHandWidth
	viewportHeight := keylistHeight - headerHeight
	m.viewport = viewport.New(viewportWidth, viewportHeight)
	m.viewport.Style = viewportStyle.Width(viewportWidth)
	m.viewport.YPosition = headerHeight
	m.setViewportContent(ctx)
}

func NewModel(data *data.Data) model {
	m := model{}

	km := ui.NewListKeyMap()
	m.data = data

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

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	var ctx context.Context = context.Background()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		util.Debug("key pressed: ", msg.String())
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			err := m.data.Close()
			if err != nil {
				fmt.Println("error closing connection: ", err)
			}
			return m, tea.Quit
		case "enter":
			m.keylist.SetItems([]list.Item{})                    // clear items
			pageSize := m.keylist.Paginator.ItemsOnPage(1000)    // estimate the page size
			m.scan = data.NewScan(m.textinput.Value(), pageSize) // initialize scan
			m.scanCh = m.data.ScanAsync(ctx, m.scan)             // start scan
			m.keylist, cmd = m.keylist.Update(msg)
			return m, tea.Batch(append(cmds, cmd)...)
		case "up", "down", "left", "?", "home", "end", "pgdown", "pgup":
			var cmd tea.Cmd
			m.keylist, cmd = m.keylist.Update(msg)
			m.resizeViews(ctx)
			return m, tea.Batch(cmd)
		case "ctrl+t", "right":
			// If on the last page and the current scan is complete,
			// then we can scan for the next page of results.
			// And ctrl+t? That's just an undocumented shortcut.
			if m.keylist.Paginator.OnLastPage() && m.scan != nil && !m.scan.Scanning() && m.scan.HasMore() {
				m.scanCh = m.data.ScanAsync(ctx, m.scan)
			}
			m.keylist, cmd = m.keylist.Update(msg)
			m.resizeViews(ctx)
			return m, tea.Batch(append(cmds, cmd)...)
		}
	case tea.WindowSizeMsg:
		// WindowSizeMsg is sent before the first render and then again every resize.
		m.windowHeight, m.windowWidth = msg.Height, msg.Width
		m.resizeViews(ctx)
		m.initialized = true
	case error:
		// handle errors like any other message
		m.viewport.SetContent(msg.Error())
		return m, nil
	}

	cmds = append(cmds, m.readAndInsert()...)

	if m.viewport.VisibleLineCount() == 0 {
		// On new searches, update the viewport with the first list item.
		m.setViewportContent(ctx)
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
			cmd := m.keylist.InsertItem(math.MaxInt, Key{*k})
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
	inputBlock := headerStyle.Copy().
		Width(LeftHandWidth() - 6).
		Align(lipgloss.Left).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			m.textinput.View(),
			m.spinnerView(),
		))
	statusBlock := headerStyle.Copy().
		Width(RightHandWidth).
		Align(lipgloss.Right).
		Render(lipgloss.JoinVertical(lipgloss.Right,
			m.data.URI(),
			fmt.Sprintf("%d keys", m.data.TotalKeys(context.Background())),
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
		markdown := m.data.Fetch(ctx, m.keylist.SelectedItem().(Key).Key)
		renderer := util.PanicOnError(glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.viewport.Width),
		))

		str := util.PanicOnError(renderer.Render(markdown))
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

///////////////////////////////////

// Key represents a Redis key, and implements [list.Item]
type Key struct {
	data.Key
}

func (k Key) String() string {
	return fmt.Sprintf("%s (%s)", k.Name, k.Datatype)
}

func (k Key) TTLString() string {
	if k.TTL == -1 {
		return "∞"
	}
	return humanize.RelTime(time.Now(), time.Now().Add(k.TTL), "", "")
}

func (k Key) SizeString() string {
	return humanize.Bytes(k.Size)
}

func (k Key) Title() string {
	typeLabel := lipgloss.NewStyle().Background(ColorForKeyType(k.Datatype)).Render(k.Datatype)
	return lipgloss.NewStyle().Width(TypeLabelWidth).Render(typeLabel) +
		lipgloss.NewStyle().Width(KeyNameWidth).Inline(true).Render(k.Name) +
		lipgloss.NewStyle().Width(TTLWidth).Render(k.TTLString()) +
		lipgloss.NewStyle().Width(SizeWidth).Render(k.SizeString())
}

func (k Key) Description() string {
	return ""
}

func (k Key) FilterValue() string {
	return k.Name
}

// //////////////////////////////////////////
// style-specific vars and funcs
// //////////////////////////////////////////

var (
	TypeLabelWidth = 10 // max is "string"
	KeyNameWidth   = 20 // assume the max to start, and adjust as keys are found
	TTLWidth       = 12 // max is "101 minutes"
	SizeWidth      = 7
	RightHandWidth = 30
)

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

func LeftHandWidth() int {
	return TypeLabelWidth + KeyNameWidth + TTLWidth + SizeWidth + 3
}

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

///////////////////////////////////////////

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	debugFlag := flag.Bool("debug", false, "Enable debug logging to the debug.log file")
	clusterFlag := flag.Bool("c", false, "Use cluster mode")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s (%s, built on %s)\n", version, commit, date)
		os.Exit(0)
	}

	if *debugFlag {
		// all calls to fmt.Println will be written to debug.log
		util.Logfile = util.PanicOnError(tea.LogToFile("debug.log", "debug"))
	}

	uri := flag.Arg(0)
	if uri == "" {
		uri = "redis://localhost:6379"
	}

	d := data.NewData(uri, *clusterFlag)
	p := tea.NewProgram(
		NewModel(d),
		tea.WithAltScreen(), // use the full size of the terminal in the alternate screen buffer
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		util.Logfile.Close()
		os.Exit(1)
	}

	util.Logfile.Close()
}
