package screen

import (
	"guessh/internal/logger"
	"guessh/internal/protocol"

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
	WaitingOpponentScreenID
	MatchResultsScreenID
)

type MatchFinishedMsg struct {
	roundsPlayed int
	roundsWon    int
}

type StartGameMsg struct{}

type mainModel struct {
	width, height   int
	matchInfo       *MatchInfo
	confirm         *bool
	screenID        ScreenID
	form            *huh.Form
	game            tea.Model
	waitingOpponent tea.Model
	matchResults    tea.Model
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
	logger.Debug("Update(%v)", msg)
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			logger.Debug("Program quitting...")
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		logger.Debug("[Update] Window resizing...")
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case StartGameMsg:
		m.screenID = StartScreenID
		m.matchInfo = NewMatchInfo()
		m.form, m.confirm = NewStartMenu(m.matchInfo)

		return m, tea.Batch(
			tea.ClearScreen,
			m.form.Init(),
		)

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
			if !*m.confirm {
				m.screenID = StartScreenID
				m.form, m.confirm = NewStartMenu(m.matchInfo)

				return m, tea.ClearScreen
			}

			switch m.matchInfo.mode {
			case protocol.MULTI_REMOTE:
				if m.matchInfo.joinExisting {
					m.screenID = GameScreenID
					m.game = NewGame(m.matchInfo)
					return m, tea.Batch(cmd, m.game.Init())
				} else {
					m.screenID = WaitingOpponentScreenID

					m.waitingOpponent = NewWaitingOpponentModel(m.matchInfo.roomID)
					return m, tea.Batch(cmd, m.waitingOpponent.Init())
				}
			case protocol.SINGLE:
				m.screenID = GameScreenID
				m.game = NewGame(m.matchInfo)
				return m, tea.Batch(cmd, m.game.Init())
			}
		}
		return m, cmd

	case GameScreenID:
		updatedModel, gameCmd := m.game.Update(msg)
		m.game = updatedModel

		return m, gameCmd

	case WaitingOpponentScreenID:
		updatedModel, waitingOpponentCmd := m.waitingOpponent.Update(msg)
		m.waitingOpponent = updatedModel
		return m, waitingOpponentCmd

	case MatchResultsScreenID:
		updatedModel, matchResultsCmd := m.matchResults.Update(msg)
		m.matchResults = updatedModel

		return m, matchResultsCmd
	}

	return m, nil
}

func (m mainModel) View() string {
	logger.Debug("View()")
	var content string

	switch m.screenID {
	case StartScreenID:
		content = contentStyle.Render(m.form.View())
	case GameScreenID:
		content = contentStyle.Render(m.game.View())
	case WaitingOpponentScreenID:
		content = contentStyle.Render(m.waitingOpponent.View())
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
