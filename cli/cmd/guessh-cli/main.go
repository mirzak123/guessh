package main

import (
	"guessh/internal/screen"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var contentStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(1, 2)

type ScreenID string

const (
	StartScreenID ScreenID = "start"
	GameScreenID  ScreenID = "game"
)

type model struct {
	width, height int
	matchInfo     *screen.MatchInfo
	confirm       *bool
	screenID      ScreenID
	form          *huh.Form
	game          tea.Model
}

func initialModel() model {
	m := model{
		screenID:  StartScreenID,
		matchInfo: screen.NewMatchInfo(),
	}
	m.form, m.confirm = screen.NewStartMenu(m.matchInfo)
	return m
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			log.Print("Program quitting...")
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		log.Print("[Update] Window resizing...")
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case screen.GameFinishedMsg:
		m.screenID = StartScreenID
		m.matchInfo = screen.NewMatchInfo()
		m.form, m.confirm = screen.NewStartMenu(m.matchInfo)
		return m, nil
	}

	switch m.screenID {

	case StartScreenID:
		updatedModel, formCmd := m.form.Update(msg)
		m.form = updatedModel.(*huh.Form)
		cmd = formCmd

		if m.screenID == StartScreenID && m.form.State == huh.StateCompleted {
			if *m.confirm {
				m.screenID = GameScreenID
				m.game = screen.NewGame(m.matchInfo)

				return m, tea.Batch(cmd, m.game.Init())
			} else {
				m.screenID = StartScreenID
				m.form, m.confirm = screen.NewStartMenu(m.matchInfo)

				return m, tea.ClearScreen
			}
		}
		return m, cmd

	case GameScreenID:
		updatedModel, gameCmd := m.game.Update(msg)
		m.game = updatedModel

		// TODO: If game state is StateMatchFinished, move back to a new form

		return m, gameCmd
	}

	return m, nil
}

func (m model) View() string {
	var content string

	switch m.screenID {
	case StartScreenID:
		content = contentStyle.Render(m.form.View())
	case GameScreenID:
		content = contentStyle.Render(m.game.View())
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func main() {
	tea.LogToFile("cli.log", "")

	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v\n", err)
	}
}
