package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
)

type TickMsg time.Time

type session struct {
	tea.Model
	_term term
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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m._term.updateHW(msg.Height, msg.Width)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case TickMsg:
		return m, fetchData()
	}
	return m, nil
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
	m := session{
		_term: term{
			user:   s.Context().User(),
			term:   pty.Term,
			width:  pty.Window.Width,
			height: pty.Window.Height,
		},
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
	s := table.New().Width(80).Headers([]string{"TERM", "WIDTH", "HEIGHT", "USER"}...).Rows([][]string{
		{m._term.term, fmt.Sprintf("%d", m._term.width), fmt.Sprintf("%d", m._term.height), m._term.user},
	}...)

	return s.Render()
}
