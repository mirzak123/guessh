package screen

import (
	"encoding/json"
	"fmt"
	"guessh/internal/client"
	"guessh/internal/game"
	"guessh/internal/logger"
	"guessh/internal/protocol"
	"guessh/internal/transport"
	"guessh/internal/ui"
	"net"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

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
	game            *gameModel
	waitingOpponent *waitingOpponentModel
	matchResults    *matchResultsModel
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
	var baseCmd tea.Cmd
	if !m.eventsPaused {
		baseCmd = transport.WaitForEvent(m.msg)
	}

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
		return m, baseCmd

	case game.StartGameIntent:
		m.screenID = StartScreenID
		m.matchInfo = game.NewMatchInfo()
		m.form, m.confirm = NewStartMenu(m.matchInfo)

		return m, tea.Batch(
			tea.ClearScreen,
			m.form.Init(),
			baseCmd,
		)

	case MatchFinishedMsg:
		m.screenID = MatchResultsScreenID
		m.matchResults = NewMatchResults(msg.roundsPlayed, msg.roundsWon)

	case RoomCreatedMsg:
		logger.Debug("setting room ID: %s", msg.roomID)
		m.matchInfo.RoomID = msg.roomID
		m.waitingOpponent.roomID = msg.roomID

	case game.CreateMatchIntent:
		m.client.CreateMatch(msg.Mode, msg.WordLen, msg.Rounds)

	case game.JoinRoom:
		m.client.JoinRoom(m.matchInfo.RoomID)

	case game.MakeGuessIntent:
		m.client.MakeGuess(msg.Guess)

	case game.PauseIntent:
		logger.Debug("Pausing Events")
		m.eventsPaused = true

	case game.ContinueIntent:
		logger.Debug("Continuing Events")
		m.eventsPaused = false

	case transport.EventMsg:
		logger.Info("New event received: %s\n", string(msg))

		msgFromEvent := m.handleEvent(msg)
		if msgFromEvent != nil {
			return m, emit(msgFromEvent)
		}
	}

	switch m.screenID {

	case StartScreenID:
		updatedModel, formCmd := m.form.Update(msg)
		m.form = updatedModel.(*huh.Form)

		if m.form.State == huh.StateCompleted {
			if !*m.confirm {
				m.screenID = StartScreenID
				m.form, m.confirm = NewStartMenu(m.matchInfo)

				return m, tea.Batch(
					tea.ClearScreen,
					formCmd,
					baseCmd,
				)
			}

			if m.matchInfo.RoomID != "" {
				m.matchInfo.JoinExisting = true
			}

			m.game = NewGame(m.matchInfo)

			switch m.matchInfo.Mode {
			case protocol.MULTI_REMOTE:
				if m.matchInfo.JoinExisting {
					m.screenID = GameScreenID
					return m, tea.Batch(
						m.game.Init(),
						baseCmd,
					)
				} else {
					m.screenID = WaitingOpponentScreenID
					m.waitingOpponent = NewWaitingOpponentModel(m.matchInfo.RoomID)

					return m, tea.Batch(
						m.game.Init(),
						m.waitingOpponent.Init(),
						baseCmd,
					)
				}
			case protocol.SINGLE:
				m.screenID = GameScreenID
				return m, tea.Batch(
					m.game.Init(),
					baseCmd,
				)
			}
		}
		return m, tea.Batch(
			formCmd,
			baseCmd,
		)

	case GameScreenID:
		_, gameCmd := m.game.Update(msg)

		return m, tea.Batch(
			gameCmd,
			baseCmd,
		)

	case WaitingOpponentScreenID:
		// BUG: ROOM_CREATED event gets passed to waiting opponent screen instead of game screen, meaning it's never processed.
		// Solution: move handleEvent to main screen and pass events to the screen that needs them
		_, waitingOpponentCmd := m.waitingOpponent.Update(msg)
		return m, tea.Batch(
			waitingOpponentCmd,
			baseCmd,
		)

	case MatchResultsScreenID:
		_, matchResultsCmd := m.matchResults.Update(msg)

		return m, tea.Batch(
			matchResultsCmd,
			baseCmd,
		)
	}

	return m, baseCmd
}

func (m mainModel) View() string {
	var content string

	switch m.screenID {
	case StartScreenID:
		content = ui.MainContentStyle.Render(m.form.View())
	case GameScreenID:
		content = ui.MainContentStyle.Render(m.game.View())
	case WaitingOpponentScreenID:
		content = ui.MainContentStyle.Render(m.waitingOpponent.View())
	case MatchResultsScreenID:
		content = ui.MainContentStyle.Render(m.matchResults.View())
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (m *mainModel) handleEvent(eventMsg transport.EventMsg) tea.Msg {
	msg := []byte(eventMsg)
	event := &protocol.EnvelopeMessage{}

	if err := json.Unmarshal(msg, event); err != nil {
		logger.Error("[handleEvent] error unmarshaling EnvelopeMessage: %s", err)
		return nil
	}

	logger.Info("[handleEvent] Event type: %s", event.Type)

	switch event.Type {

	case protocol.MATCH_STARTED:
		m.game.state = game.StateMatchStarted
		m.screenID = GameScreenID

		matchStartedEvent := &protocol.MatchStartedMessage{}

		if err := json.Unmarshal(msg, matchStartedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoundStartedMessage: %v", err)
			return nil
		}

		m.matchInfo.WordLen = matchStartedEvent.WordLength
		m.matchInfo.TotalRounds = matchStartedEvent.Rounds
		m.matchInfo.RawTotalRounds = fmt.Sprintf("%d", matchStartedEvent.Rounds)

	case protocol.ROUND_STARTED:
		m.matchInfo.CurrentRound++
		m.game.roundInfo = game.NewRoundInfo()

		roundStartedEvent := &protocol.RoundStartedMessage{}

		if err := json.Unmarshal(msg, roundStartedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoundStartedMessage: %v", err)
			return nil
		}
		m.matchInfo.MaxAttempts = roundStartedEvent.MaxAttempts
		m.game.guesses = nil

	case protocol.WAIT_GUESS:
		m.game.state = game.StateWaitGuess
		m.game.input.Focus()

	case protocol.WAIT_OPPONENT_GUESS:
		m.game.state = game.StateWaitOpponentGuess
		m.game.input.Blur()

	case protocol.WAIT_OPPONENT_JOIN:
		m.game.state = game.StateWaitOpponentJoin

	case protocol.GUESS_RESULT:
		guessResultEvent := &protocol.GuessResultMessage{}

		if err := json.Unmarshal(msg, guessResultEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling GuessResultMessage: %v", err)
			return nil
		}

		m.game.guesses = append(m.game.guesses, protocol.NewGuess(guessResultEvent.Guess, guessResultEvent.Feedback))

	case protocol.ROUND_FINISHED:
		roundFinishedEvent := &protocol.RoundFinishedMessage{}

		if err := json.Unmarshal(msg, roundFinishedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoundFinishedMessage: %v", err)
			return nil
		}

		m.game.state = game.StateRoundFinished
		m.game.roundInfo.Word = roundFinishedEvent.Word
		m.game.roundInfo.Success = roundFinishedEvent.Success
		m.game.input.Blur()

		if roundFinishedEvent.Success {
			m.matchInfo.RoundsWon++
		}

		return game.PauseIntent{}

	case protocol.MATCH_FINISHED:
		m.game.state = game.StateMatchFinished
		return MatchFinishedMsg{
			roundsPlayed: m.matchInfo.TotalRounds,
			roundsWon:    m.matchInfo.RoundsWon,
		}

	case protocol.ROOM_CREATED:
		logger.Debug("Processing ROOM_CREATED event")
		roomCreatedEvent := &protocol.RoomCreatedMessage{}

		if err := json.Unmarshal(msg, roomCreatedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoomCreatedMessage: %v", err)
			return nil
		}

		return RoomCreatedMsg{
			roomID: roomCreatedEvent.RoomID,
		}

	}

	return nil
}
