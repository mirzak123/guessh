package screen

import (
	"encoding/json"
	"errors"
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
	width, height   int
	client          *client.Client
	event           chan transport.EventMsg
	eventBuffer     []transport.EventMsg
	eventsPaused    bool
	screenID        ScreenID
	matchInfo       *game.MatchInfo
	confirm         *bool
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
		event:     make(chan transport.EventMsg),
	}
	m.form, m.confirm = NewStartMenu(m.matchInfo)
	return m
}

func (m mainModel) Init() tea.Cmd {

	return tea.Batch(
		tea.EnterAltScreen,
		transport.ListenForActivity(m.client.Conn, m.event),
		transport.WaitForEvent(m.event),
	)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

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
		return m, tea.Batch(cmds...)

	case game.StartGameIntent:
		m.screenID = StartScreenID
		m.matchInfo = game.NewMatchInfo()
		m.form, m.confirm = NewStartMenu(m.matchInfo)

		m.eventsPaused = false

		return m, tea.Batch(tea.ClearScreen, m.form.Init(), transport.WaitForEvent(m.event))

	case MatchFinishedMsg:
		m.screenID = MatchResultsScreenID
		m.matchResults = NewMatchResults(msg.roundsPlayed, msg.roundsWon)

	case RoomCreatedMsg:
		logger.Debug("Setting room ID: %s", msg.roomID)
		m.matchInfo.RoomID = msg.roomID
		m.waitingOpponent.roomID = msg.roomID

	case game.CreateMatchIntent:
		m.client.CreateMatch(msg.Mode, msg.WordLen, msg.Rounds, msg.PlayerName)

	case game.JoinRoomIntent:
		m.client.JoinRoom(msg.RoomId, msg.PlayerName)

	case game.MakeGuessIntent:
		m.client.MakeGuess(msg.Guess)

	case game.PauseIntent:
		logger.Debug("Pausing Events")
		m.eventsPaused = true

	case game.ContinueIntent:
		logger.Debug("Continuing events, flushing buffer")
		m.eventsPaused = false

		for _, bufferedEvent := range m.eventBuffer {
			msgFromEvent := m.handleEvent(bufferedEvent)
			if msgFromEvent != nil {
				cmds = append(cmds, emit(msgFromEvent))
			}
		}
		m.eventBuffer = nil

	case game.TypingIntent:
		m.client.Typing(msg.Value)

	case transport.EventMsg:
		cmds = append(cmds, transport.WaitForEvent(m.event))

		if m.eventsPaused {
			logger.Debug("UI paused, buffering event")
			m.eventBuffer = append(m.eventBuffer, msg)
		} else {
			msgFromEvent := m.handleEvent(msg)
			if msgFromEvent != nil {
				cmds = append(cmds, emit(msgFromEvent))
			}
		}
	}

	switch m.screenID {

	case StartScreenID:
		updatedModel, formCmd := m.form.Update(msg)
		m.form = updatedModel.(*huh.Form)

		cmds = append(cmds, formCmd)

		if m.form.State == huh.StateCompleted {
			if !*m.confirm {
				m.screenID = StartScreenID
				m.form, m.confirm = NewStartMenu(m.matchInfo)

				return m, tea.Batch(tea.ClearScreen, formCmd)
			}

			if m.matchInfo.RoomID != "" {
				m.matchInfo.JoinExisting = true
			}

			m.game = NewGame(m.matchInfo)

			switch m.matchInfo.Mode {
			case protocol.MULTI_REMOTE:
				if m.matchInfo.JoinExisting {
					m.screenID = GameScreenID
					cmds = append(cmds, m.game.Init())
				} else {
					m.screenID = WaitingOpponentScreenID
					m.waitingOpponent = NewWaitingOpponentModel(m.matchInfo.RoomID)
					cmds = append(cmds, m.game.Init(), m.waitingOpponent.Init())
				}
			case protocol.SINGLE:
				m.screenID = GameScreenID
				cmds = append(cmds, m.game.Init())
			}
		}

	case GameScreenID:
		_, gameCmd := m.game.Update(msg)
		cmds = append(cmds, gameCmd)

	case WaitingOpponentScreenID:
		_, waitingOpponentCmd := m.waitingOpponent.Update(msg)
		cmds = append(cmds, waitingOpponentCmd)

	case MatchResultsScreenID:
		_, matchResultsCmd := m.matchResults.Update(msg)
		cmds = append(cmds, matchResultsCmd)
	}

	return m, tea.Batch(cmds...)
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
		matchStartedEvent := &protocol.MatchStartedMessage{}
		if err := json.Unmarshal(msg, matchStartedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoundStartedMessage: %v", err)
			return nil
		}

		m.game.state = game.StateMatchStarted
		m.screenID = GameScreenID

		m.matchInfo.WordLen = matchStartedEvent.WordLength
		m.matchInfo.TotalRounds = matchStartedEvent.Rounds
		m.matchInfo.RawTotalRounds = fmt.Sprintf("%d", matchStartedEvent.Rounds)

		m.game.input.CharLimit = m.matchInfo.WordLen
		m.game.input.Width = m.matchInfo.WordLen

	case protocol.ROUND_STARTED:
		roundStartedEvent := &protocol.RoundStartedMessage{}
		if err := json.Unmarshal(msg, roundStartedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoundStartedMessage: %v", err)
			return nil
		}

		m.game.roundInfo = game.NewRoundInfo()

		m.matchInfo.MaxAttempts = roundStartedEvent.MaxAttempts
		m.matchInfo.CurrentRound = roundStartedEvent.RoundNumber
		m.game.guesses = nil

	case protocol.WAIT_GUESS:
		m.game.state = game.StateWaitGuess
		m.game.input.Focus()
		m.game.input.SetValue("")

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
			logger.Debug("Incrementing correct guess count: %d", m.matchInfo.RoundsWon)
			m.matchInfo.RoundsWon++
		} else {
			logger.Debug("Round not successful, not incrementing") // TODO: Remove
		}

		return game.PauseIntent{}

	case protocol.MATCH_FINISHED:
		m.game.state = game.StateMatchFinished
		return MatchFinishedMsg{
			roundsPlayed: m.matchInfo.TotalRounds,
			roundsWon:    m.matchInfo.RoundsWon,
		}

	case protocol.ROOM_CREATED:
		roomCreatedEvent := &protocol.RoomCreatedMessage{}
		if err := json.Unmarshal(msg, roomCreatedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoomCreatedMessage: %v", err)
			return nil
		}

		return RoomCreatedMsg{
			roomID: roomCreatedEvent.RoomID,
		}

	case protocol.ROOM_JOIN_FAILED:
		roomJoinFailedEvent := &protocol.RoomJoinFailedMessage{}
		if err := json.Unmarshal(msg, roomJoinFailedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoomJoinFailedMessage: %v", err)
			return nil
		}

		m.screenID = StartScreenID
		m.form, m.confirm = NewStartMenu(m.matchInfo)
		m.game.matchInfo.RoomValidationError = errors.New(roomJoinFailedEvent.Reason)

		// navigate to the RoomID input field
		m.form.NextGroup()
		m.form.NextGroup()
		m.form.NextField()

		// simulate a key press to register the error message
		enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
		_, formCmd := m.form.Update(enterMsg)
		return tea.Batch(tea.ClearScreen, formCmd)

	case protocol.OPPONENT_TYPING:
		opponentTypingEvent := &protocol.OpponentTypingMessage{}
		if err := json.Unmarshal(msg, opponentTypingEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling OpponentTypingMessage: %v", err)
			return nil
		}
		m.game.input.SetValue(opponentTypingEvent.Value)

	default:
		logger.Info("Ignoring event: %s", event.Type)
	}

	return nil
}
