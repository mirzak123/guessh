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
	RequestRematchScreenID
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
	connected     bool
	event         chan transport.EventMsg
	eventBuffer   []transport.EventMsg
	eventsPaused  bool
	flushing      bool
	matchInfo     *game.MatchInfo
	confirm       *bool

	screenID              ScreenID
	startMenu             *huh.Form
	game                  *gameModel
	waitingOpponentScreen *waitingOpponentModel
	requestRematchScreen  *requestRematchModel
	matchResultsScreen    *matchResultsModel
	serverDownScreen      *huh.Form
}

func InitialModel() mainModel {
	m := mainModel{
		screenID:         StartScreenID,
		matchInfo:        game.NewMatchInfo(),
		event:            make(chan transport.EventMsg),
		serverDownScreen: NewServerDownForm(),
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
		m.matchResultsScreen = NewMatchResults(
			m.matchInfo.Mode,
			msg.roundsPlayed,
			msg.roundOutcomes,
			msg.matchOutcome,
			msg.opponentLeft,
		)

	case RoomCreatedMsg:
		logger.Debug("Setting room ID: %s", msg.roomID)
		m.matchInfo.RoomID = msg.roomID
		m.waitingOpponentScreen.roomID = msg.roomID

	case game.CreateMatchIntent:
		m.client.CreateMatch(msg.Mode, msg.WordLen, msg.Rounds, msg.PlayerName)

	case game.JoinRoomIntent:
		m.client.JoinRoom(msg.RoomId, msg.PlayerName)

	case game.RequestRematchIntent:
		m.client.RequestRematch()
		m.screenID = RequestRematchScreenID
		m.requestRematchScreen = NewRequestRematchModel(m.matchInfo.OpponentName)
		cmds = append(cmds, m.requestRematchScreen.Init())

	case game.DenyRematchIntent:
		m.client.DenyRematch()
		cmds = append(cmds, emit(game.StartGameIntent{}))

	case game.MakeGuessIntent:
		m.client.MakeGuess(msg.Guess)

	case game.ContinueIntent:
		logger.Debug("Continuing events, flushing buffer")
		m.eventsPaused = false
		m.flushing = true

		for _, bufferedEvent := range m.eventBuffer {
			msgFromEvent := m.handleEvent(bufferedEvent)
			if msgFromEvent != nil {
				cmds = append(cmds, emit(msgFromEvent))
			}
		}

		m.flushing = false
		m.eventBuffer = nil

	case game.LeaveMatchIntent:
		m.client.LeaveMatch()
		cmds = append(cmds, emit(game.StartGameIntent{}))

	case game.TypingIntent:
		if m.matchInfo.Mode == protocol.MULTI_REMOTE {
			m.client.Typing(msg.Value)
		}

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
		logger.Debug("[%s] event buffer: %v", m.matchInfo.PlayerName, m.eventBuffer)
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
					m.waitingOpponentScreen = NewWaitingOpponentModel()
					cmds = append(cmds, m.game.Init(), m.waitingOpponentScreen.Init())
				}
			case protocol.SINGLE:
				m.screenID = GameScreenID
				cmds = append(cmds, m.game.Init())
			}
		}

	case ServerDownScreenID:
		updatedModel, formCmd := m.serverDownScreen.Update(msg)
		m.serverDownScreen = updatedModel.(*huh.Form)

		if m.serverDownScreen.State == huh.StateCompleted {
			return m, tea.Quit
		}

		return m, formCmd

	case GameScreenID:
		_, gameCmd := m.game.Update(msg)
		cmds = append(cmds, gameCmd)

	case WaitingOpponentScreenID:
		_, waitingOpponentCmd := m.waitingOpponentScreen.Update(msg)
		cmds = append(cmds, waitingOpponentCmd)

		if m.waitingOpponentScreen.form.State == huh.StateCompleted {
			cmds = append(cmds, emit(game.LeaveMatchIntent{}))
		}

	case RequestRematchScreenID:
		_, rematchRequestedCmd := m.requestRematchScreen.Update(msg)
		cmds = append(cmds, rematchRequestedCmd)

		if m.requestRematchScreen.form.State == huh.StateCompleted {
			cmds = append(cmds, emit(game.DenyRematchIntent{}))
		}

	case MatchResultsScreenID:
		_, matchResultsCmd := m.matchResultsScreen.Update(msg)
		cmds = append(cmds, matchResultsCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	var content string

	switch m.screenID {
	case StartScreenID:
		content = m.startMenu.View()
	case ServerDownScreenID:
		content = m.serverDownScreen.View()
	case GameScreenID:
		content = m.game.View()
	case WaitingOpponentScreenID:
		content = m.waitingOpponentScreen.View()
	case RequestRematchScreenID:
		content = m.requestRematchScreen.View()
	case MatchResultsScreenID:
		content = m.matchResultsScreen.View()
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		ui.MainContentStyle.Render(content),
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
		m.game.input.SetValue("")

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
		if !m.flushing {
			m.eventsPaused = true
		}
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

	case protocol.OPPONENT_DENIED_REMATCH:
		m.requestRematchScreen.opponentDeniedRematch = true

	default:
		logger.Info("Ignoring event: %s", event.Type)
	}

	return nil
}
