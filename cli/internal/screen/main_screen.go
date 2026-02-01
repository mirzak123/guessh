package screen

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var contentStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(1, 2)

type ScreenID int

const (
	StartScreenID ScreenID = iota
	GameScreenID
	MatchResultsScreenID
)

type MatchFinishedMsg struct {
	roundsPlayed int
	roundsWon    int
}
type StartGameMsg struct{}

type mainModel struct {
	width, height int
	matchInfo     *MatchInfo
	confirm       *bool
	screenID      ScreenID
	form          *huh.Form
	game          tea.Model
	matchResults  tea.Model
}

func InitialModel() mainModel {
	m := mainModel{
		screenID:  StartScreenID,
		matchInfo: NewMatchInfo(),
	}
	m.form, m.confirm = NewStartMenu(m.matchInfo)
	return m
}

func (m mainModel) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case StartGameMsg:
		m.screenID = StartScreenID
		m.matchInfo = NewMatchInfo()
		m.form, m.confirm = NewStartMenu(m.matchInfo)
		return m, nil

	case MatchFinishedMsg:
		m.screenID = MatchResultsScreenID
		m.matchResults = NewMatchResults(msg.roundsPlayed, msg.roundsWon)
	}

	switch m.screenID {

	case StartScreenID:
		updatedModel, formCmd := m.form.Update(msg)
		m.form = updatedModel.(*huh.Form)
		cmd = formCmd

		if m.screenID == StartScreenID && m.form.State == huh.StateCompleted {
			if *m.confirm {
				m.screenID = GameScreenID
				m.game = NewGame(m.matchInfo)

				return m, tea.Batch(cmd, m.game.Init())
			} else {
				m.screenID = StartScreenID
				m.form, m.confirm = NewStartMenu(m.matchInfo)

				return m, tea.ClearScreen
			}
		}
		return m, cmd

	case GameScreenID:
		updatedModel, gameCmd := m.game.Update(msg)
		m.game = updatedModel

		return m, gameCmd

	case MatchResultsScreenID:
		updatedModel, matchResultsCmd := m.matchResults.Update(msg)
		m.matchResults = updatedModel

		return m, matchResultsCmd
	}

	return m, nil
}

func (m mainModel) View() string {
	var content string

	switch m.screenID {
	case StartScreenID:
		content = contentStyle.Render(m.form.View())
	case GameScreenID:
		content = contentStyle.Render(m.game.View())
	case MatchResultsScreenID:
		content = contentStyle.Render(m.matchResults.View())
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
