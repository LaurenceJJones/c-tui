package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
)

type TickMsg time.Time

// Just a generic tea.Model to demo terminal information of ssh.
type model struct {
	term      string
	width     int
	height    int
	txtStyle  lipgloss.Style
	quitStyle lipgloss.Style
}

func (m model) Init() tea.Cmd {
	return fetchData()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
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
	renderer := bm.MakeRenderer(s)
	m := model{
		term:      pty.Term,
		width:     pty.Window.Width,
		height:    pty.Window.Height,
		txtStyle:  renderer.NewStyle().Foreground(lipgloss.Color("10")),
		quitStyle: renderer.NewStyle().Foreground(lipgloss.Color("8")),
	}
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

func fetchData() tea.Cmd {
	// TODO: fetch data every 10 seconds
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m model) View() string {
	s := table.New().Width(80).Headers([]string{"TERM", "WIDTH", "HEIGHT"}...).Rows([][]string{
		{m.term, fmt.Sprintf("%d", m.width), fmt.Sprintf("%d", m.height)},
	}...)

	return s.Render()
}
