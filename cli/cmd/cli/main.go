package main

import (
	"guessh/cmd/cli/screen"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var formStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(1, 2)

type ScreenType string

const (
	StartScreenType ScreenType = "start"
	GameScreenType  ScreenType = "game"
)

type model struct {
	width, height int
	screen        ScreenType
	form          *huh.Form
	game          *tea.Model
}

func initialModel() model {
	return model{
		screen: StartScreenType,
		form:   screen.NewStartMenu(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Print("[Update] Updating model...")
	log.Print("StartGame:", screen.StartGame)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		log.Print("[Update] KeyMsg")
		switch msg.Type {
		case tea.KeyCtrlC:
			log.Print("Quitting...")
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		log.Print("[Update] Window resizing...")
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	log.Print("[Update] Calling form.Update()")
	updatedModel, cmd := m.form.Update(msg)
	m.form = updatedModel.(*huh.Form)

	if m.screen == StartScreenType && m.form.State == huh.StateCompleted {
		if screen.StartGame {
			m.screen = GameScreenType
		} else {
			m.screen = StartScreenType
			screen.StartGame = false
			m.form = screen.NewStartMenu()

			return m, tea.ClearScreen
		}
	}
	return m, cmd
}

func (m model) View() string {
	var content string

	switch m.screen {
	case StartScreenType:
		content = formStyle.Render(m.form.View())
	case GameScreenType:
		content = "Game started"
		// content = formStyle.Render(m.form.View())
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
