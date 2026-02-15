package screen

import (
	"guessh/internal/client"
	"guessh/internal/game"
	"guessh/internal/logger"
	"guessh/internal/protocol"
	"guessh/internal/transport"
	"net"
	"os"

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

type RoomCreatedMsg struct {
	roomID string
}

type mainModel struct {
	client          *client.Client
	msg             chan transport.EventMsg
	width, height   int
	matchInfo       *game.MatchInfo
	confirm         *bool
	eventsPaused    bool
	screenID        ScreenID
	form            *huh.Form
	game            tea.Model
	waitingOpponent tea.Model
	matchResults    tea.Model
}

func InitialModel() mainModel {
	conn, err := net.Dial("tcp", "localhost:2480")
	if err != nil {
		logger.Error("net.Dial error: %v", err)
		os.Exit(1)
	}

	c := client.NewClient(conn)

	m := mainModel{
		screenID:  StartScreenID,
		matchInfo: game.NewMatchInfo(),
		client:    c,
		msg:       make(chan transport.EventMsg),
	}
	m.form, m.confirm = NewStartMenu(m.matchInfo)
	return m
}

func (m mainModel) Init() tea.Cmd {

	return tea.Batch(
		tea.EnterAltScreen,
		transport.ListenForActivity(m.client.Conn, m.msg),
		transport.WaitForEvent(m.msg),
	)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case game.StartGameIntent:
		m.screenID = StartScreenID
		m.matchInfo = game.NewMatchInfo()
		m.form, m.confirm = NewStartMenu(m.matchInfo)

		return m, tea.Batch(
			tea.ClearScreen,
			m.form.Init(),
		)

	case MatchFinishedMsg:
		m.screenID = MatchResultsScreenID
		m.matchResults = NewMatchResults(msg.roundsPlayed, msg.roundsWon)

	case RoomCreatedMsg:
		m.matchInfo.RoomID = msg.roomID

	case game.CreateMatchIntent:
		m.client.CreateMatch(msg.Mode, msg.WordLen, msg.Rounds)

	case game.MakeGuessIntent:
		m.client.MakeGuess(msg.Guess)

	case game.PauseIntent:
		m.eventsPaused = true

	case game.ContinueIntent:
		m.eventsPaused = false
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

			m.game = NewGame(m.matchInfo)

			switch m.matchInfo.Mode {
			case protocol.MULTI_REMOTE:
				if m.matchInfo.JoinExisting {
					m.screenID = GameScreenID
					return m, tea.Batch(cmd, m.game.Init())
				} else {
					m.screenID = WaitingOpponentScreenID
					m.waitingOpponent = NewWaitingOpponentModel(m.matchInfo.RoomID)

					return m, tea.Batch(cmd, m.game.Init(), m.waitingOpponent.Init())
				}
			case protocol.SINGLE:
				m.screenID = GameScreenID
				return m, tea.Batch(cmd, m.game.Init())
			}
		}
		return m, cmd

	case GameScreenID:
		logger.Debug("calling game update")
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

	if m.eventsPaused {
		return m, nil
	}
	return m, transport.WaitForEvent(m.msg)
}

func (m mainModel) View() string {
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
