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
	ServerDownScreenID
)

type MatchFinishedMsg struct {
	roundsPlayed  int
	roundOutcomes []*protocol.Outcome
	matchOutcome  protocol.Outcome
	opponentLeft  bool
}

type RoomCreatedMsg struct {
	roomID string
}

type mainModel struct {
	width, height int
	client        *client.Client
	event         chan transport.EventMsg
	eventBuffer   []transport.EventMsg
	eventsPaused  bool
	screenID      ScreenID
	matchInfo     *game.MatchInfo
	confirm       *bool
	connected     bool

	startMenu       *huh.Form
	game            *gameModel
	waitingOpponent *waitingOpponentModel
	matchResults    *matchResultsModel
	serverDown      *huh.Form
}

func InitialModel() mainModel {
	m := mainModel{
		screenID:   StartScreenID,
		matchInfo:  game.NewMatchInfo(),
		event:      make(chan transport.EventMsg),
		serverDown: NewServerDownForm(),
	}

	conn, err := net.Dial("tcp", "localhost:2480")
	if err != nil {
		logger.Error("net.Dial error: %v", err)
		m.screenID = ServerDownScreenID
	} else {
		m.connected = true
	}

	m.client = client.NewClient(conn)
	m.startMenu, m.confirm = NewStartMenu(m.matchInfo)
	return m
}

func (m mainModel) Init() tea.Cmd {

	if !m.connected {
		return tea.EnterAltScreen
	}

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
		m.startMenu, m.confirm = NewStartMenu(m.matchInfo)

		m.eventsPaused = false

		return m, tea.Batch(tea.ClearScreen, m.startMenu.Init())

	case MatchFinishedMsg:
		m.screenID = MatchResultsScreenID
		m.matchResults = NewMatchResults(
			m.matchInfo.Mode,
			msg.roundsPlayed,
			msg.roundOutcomes,
			msg.matchOutcome,
			msg.opponentLeft,
		)

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

	case game.LeaveMatchIntent:
		m.client.LeaveMatch()
		cmds = append(cmds, emit(game.StartGameIntent{}))

	case game.TypingIntent:
		m.client.Typing(msg.Value)

	case transport.ServerDisconnectedMsg:
		m.screenID = ServerDownScreenID

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
		updatedModel, formCmd := m.startMenu.Update(msg)
		m.startMenu = updatedModel.(*huh.Form)

		cmds = append(cmds, formCmd)

		if m.startMenu.State == huh.StateCompleted {
			if !*m.confirm {
				m.screenID = StartScreenID
				m.startMenu, m.confirm = NewStartMenu(m.matchInfo)

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
					m.waitingOpponent = NewWaitingOpponentModel()
					cmds = append(cmds, m.game.Init(), m.waitingOpponent.Init())
				}
			case protocol.SINGLE:
				m.screenID = GameScreenID
				cmds = append(cmds, m.game.Init())
			}
		}

	case ServerDownScreenID:
		updatedModel, formCmd := m.serverDown.Update(msg)
		m.serverDown = updatedModel.(*huh.Form)

		if m.serverDown.State == huh.StateCompleted {
			return m, tea.Quit
		}

		return m, formCmd

	case GameScreenID:
		_, gameCmd := m.game.Update(msg)
		cmds = append(cmds, gameCmd)

	case WaitingOpponentScreenID:
		_, waitingOpponentCmd := m.waitingOpponent.Update(msg)
		cmds = append(cmds, waitingOpponentCmd)

		if m.waitingOpponent.form.State == huh.StateCompleted {
			cmds = append(cmds, emit(game.LeaveMatchIntent{}))
		}

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
		content = ui.MainContentStyle.Render(m.startMenu.View())
	case ServerDownScreenID:
		content = ui.MainContentStyle.Render(m.serverDown.View())
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
	event := &protocol.EnvelopeEvent{}

	if err := json.Unmarshal(msg, event); err != nil {
		logger.Error("[handleEvent] error unmarshaling EnvelopeEvent: %s", err)
		return nil
	}

	logger.Info("[handleEvent] Event type: %s", event.Type)

	switch event.Type {

	case protocol.MATCH_STARTED:
		matchStartedEvent := &protocol.MatchStartedEvent{}
		if err := json.Unmarshal(msg, matchStartedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling MatchStartedEvent: %v", err)
			return nil
		}

		m.game.state = game.StateMatchStarted
		m.screenID = GameScreenID

		m.matchInfo.WordLen = matchStartedEvent.WordLength
		m.matchInfo.TotalRounds = matchStartedEvent.Rounds
		m.matchInfo.OpponentName = matchStartedEvent.OpponentName

		m.matchInfo.RawTotalRounds = fmt.Sprintf("%d", matchStartedEvent.Rounds)
		m.matchInfo.RoundOutcomes = make([]*protocol.Outcome, matchStartedEvent.Rounds)

		m.game.input.CharLimit = m.matchInfo.WordLen
		m.game.input.Width = m.matchInfo.WordLen

	case protocol.ROUND_STARTED:
		roundStartedEvent := &protocol.RoundStartedEvent{}
		if err := json.Unmarshal(msg, roundStartedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoundStartedEvent: %v", err)
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
		guessResultEvent := &protocol.GuessResultEvent{}
		if err := json.Unmarshal(msg, guessResultEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling GuessResultEvent: %v", err)
			return nil
		}

		m.game.guesses = append(m.game.guesses, protocol.NewGuess(guessResultEvent.Guess, guessResultEvent.Feedback))

	case protocol.ROUND_FINISHED:
		roundFinishedEvent := &protocol.RoundFinishedEvent{}
		if err := json.Unmarshal(msg, roundFinishedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoundFinishedEvent: %v", err)
			return nil
		}

		m.game.state = game.StateRoundFinished
		m.game.roundInfo.Word = roundFinishedEvent.Word
		m.game.roundInfo.Success = roundFinishedEvent.Outcome != protocol.OUTCOME_NONE
		m.game.input.Blur()

		logger.Debug("Round outcomes: %v", m.matchInfo.RoundOutcomes)
		m.matchInfo.RoundOutcomes[m.matchInfo.CurrentRound-1] = &roundFinishedEvent.Outcome

		logger.Debug("Pausing events...")
		m.eventsPaused = true
		return nil

	case protocol.MATCH_FINISHED:
		matchFinishedEvent := &protocol.MatchFinishedEvent{}
		if err := json.Unmarshal(msg, matchFinishedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling MatchFinishedEvent: %v", err)
			return nil
		}

		m.game.state = game.StateMatchFinished

		return MatchFinishedMsg{
			roundsPlayed:  m.matchInfo.TotalRounds,
			roundOutcomes: m.matchInfo.RoundOutcomes,
			matchOutcome:  matchFinishedEvent.Outcome,
			opponentLeft:  matchFinishedEvent.OpponentLeft,
		}

	case protocol.ROOM_CREATED:
		roomCreatedEvent := &protocol.RoomCreatedEvent{}
		if err := json.Unmarshal(msg, roomCreatedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoomCreatedEvent: %v", err)
			return nil
		}

		return RoomCreatedMsg{
			roomID: roomCreatedEvent.RoomID,
		}

	case protocol.ROOM_JOIN_FAILED:
		roomJoinFailedEvent := &protocol.RoomJoinFailedEvent{}
		if err := json.Unmarshal(msg, roomJoinFailedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoomJoinFailedEvent: %v", err)
			return nil
		}

		m.screenID = StartScreenID
		m.startMenu, m.confirm = NewStartMenu(m.matchInfo)
		m.game.matchInfo.RoomValidationError = errors.New(roomJoinFailedEvent.Reason)

		// navigate to the RoomID input field
		m.startMenu.NextGroup()
		m.startMenu.NextGroup()
		m.startMenu.NextField()

		// simulate a key press to register the error message
		enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
		_, formCmd := m.startMenu.Update(enterMsg)
		return tea.Batch(tea.ClearScreen, formCmd)

	case protocol.OPPONENT_TYPING:
		opponentTypingEvent := &protocol.OpponentTypingEvent{}
		if err := json.Unmarshal(msg, opponentTypingEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling OpponentTypingEvent: %v", err)
			return nil
		}
		m.game.input.SetValue(opponentTypingEvent.Value)

	default:
		logger.Info("Ignoring event: %s", event.Type)
	}

	return nil
}
