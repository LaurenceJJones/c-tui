package main

import (
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type TickMsg time.Time

type session struct {
	tea.Model
	_term         term
	DecisionTable table.Model
}

type term struct {
	term   string
	width  int
	height int
	user   string
}

func (t *term) updateHW(w, h int) {
	t.width = w
	t.height = h
}

func (m session) Init() tea.Cmd {
	return fetchData()
}

func (m session) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m._term.updateHW(msg.Height, msg.Width)
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.DecisionTable.Focused() {
				m.DecisionTable.Blur()
			} else {
				m.DecisionTable.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case TickMsg:
		return m, fetchData()
	}
	m.DecisionTable, cmd = m.DecisionTable.Update(msg)
	return m, cmd
}

func AppMiddleware() wish.Middleware {
	return bm.Middleware(teaHandler)
}

// You can wire any Bubble Tea model up to the middleware with a function that
// handles the incoming ssh.Session. Here we just grab the terminal info and
// pass it to the new model. You can also return tea.ProgramOptions (such as
// tea.WithAltScreen) on a session by session basis.
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, active := s.Pty()
	if !active {
		wish.Fatalln(s, "no active terminal, skipping")
		return nil, nil
	}
	columnWidth := pty.Window.Width / 2 / 4
	m := session{
		_term: term{
			user:   s.Context().User(),
			term:   pty.Term,
			width:  pty.Window.Width,
			height: pty.Window.Height,
		},
		DecisionTable: table.New(
			table.WithColumns([]table.Column{
				{Title: "id", Width: columnWidth},
				{Title: "type", Width: columnWidth},
				{Title: "value", Width: columnWidth},
				{Title: "duration", Width: columnWidth},
			}),
			table.WithRows([]table.Row{
				{"1", "type1", "value1", "duration1"},
				{"1", "type1", "value1", "duration1"},
				{"1", "type1", "value1", "duration1"},
				{"1", "type1", "value1", "duration1"},
				{"1", "type1", "value1", "duration1"},
				{"1", "type1", "value1", "duration1"},
			}),
			table.WithFocused(true),
			table.WithHeight(pty.Window.Height/2),
		),
	}
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

func fetchData() tea.Cmd {
	// TODO: fetch data every 10 seconds
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m session) View() string {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	if m.DecisionTable.Focused() {
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(true)
	}
	m.DecisionTable.SetStyles(s)
	return baseStyle.Render(m.DecisionTable.View()) + "\n"
}
