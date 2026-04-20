package screen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"guessh/internal/client"
	"guessh/internal/game"
	"guessh/internal/logger"
	"guessh/internal/protocol"
	"guessh/internal/transport"
	"guessh/internal/ui"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type ScreenID int

const (
	StartMenuScreenID ScreenID = iota
	GameConfigScreenID
	GameScreenID
	WaitingOpponentScreenID
	RequestRematchScreenID
	MatchResultsScreenID
	ServerDownScreenID
)

var highPriorityEvents = []protocol.EventType{protocol.WAIT_GUESS, protocol.WAIT_OPPONENT_GUESS}

type MatchFinishedMsg struct {
	roundsPlayed int
	roundPoints  []int
	matchOutcome protocol.Outcome
	opponentLeft bool
}

type RoomCreatedMsg struct {
	roomID string
}

type forceRenderMsg struct{}

type mainModel struct {
	width, height int
	sshContext    context.Context
	client        *client.Client
	connected     bool
	event         chan transport.EventMsg
	eventBuffer   []transport.EventMsg
	eventsPaused  bool
	flushing      bool
	matchInfo     *game.MatchInfo
	confirm       *bool
	hoveredMode   protocol.GameMode

	screenID              ScreenID
	startMenuScreen       *startMenuModel
	gameConfigMenu        *huh.Form
	game                  *gameModel
	waitingOpponentScreen *waitingOpponentModel
	requestRematchScreen  *requestRematchModel
	matchResultsScreen    *matchResultsModel
	serverDownScreen      *huh.Form
}

func InitialModel() *mainModel {
	game.EnsureDictionariesLoaded()

	m := &mainModel{
		screenID:             StartMenuScreenID,
		matchInfo:            game.NewMatchInfo(),
		event:                make(chan transport.EventMsg),
		startMenuScreen:      NewStartMenu(),
		serverDownScreen:     NewServerDownForm(),
		requestRematchScreen: NewRequestRematchModel(),
	}

	conn, err := transport.Connect()
	if err != nil {
		logger.Error("Error while connecting to game server: %v", err)
		m.screenID = ServerDownScreenID
	} else {
		m.connected = true
	}

	m.client = client.NewClient(conn)
	m.gameConfigMenu, m.confirm = NewGameConfigMenu(m.matchInfo, &m.hoveredMode)
	return m
}

func (m *mainModel) Init() tea.Cmd {
	var cmds []tea.Cmd

	ctx := m.sshContext
	if ctx == nil {
		ctx = context.Background()
	}

	cmds = append(cmds,
		tea.EnterAltScreen,
		waitForContextDone(ctx),
	)

	if m.connected {
		cmds = append(cmds,
			transport.ListenForActivity(m.client.Conn, m.event),
			transport.WaitForEvent(m.event),
		)
	}

	return tea.Batch(cmds...)
}

func (m *mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case game.StartMenuIntent:
		m.screenID = StartMenuScreenID

	case game.PlayGameIntent:
		m.screenID = GameConfigScreenID
		m.matchInfo = game.NewMatchInfo()
		m.gameConfigMenu, m.confirm = NewGameConfigMenu(m.matchInfo, &m.hoveredMode)

		m.eventsPaused = false

		return m, tea.Batch(tea.ClearScreen, m.gameConfigMenu.Init())

	case MatchFinishedMsg:
		m.screenID = MatchResultsScreenID
		m.matchResultsScreen = NewMatchResults(
			m.matchInfo.Mode,
			m.matchInfo.Format,
			msg.roundsPlayed,
			msg.roundPoints,
			msg.matchOutcome,
			m.matchInfo.PlayerName,
			m.matchInfo.OpponentName,
			msg.opponentLeft,
		)

	case RoomCreatedMsg:
		logger.Debug("Setting room ID: %s", msg.roomID)
		m.matchInfo.RoomID = msg.roomID
		m.waitingOpponentScreen.roomID = msg.roomID

	case game.CreateMatchIntent:
		m.client.CreateMatch(msg.Mode, msg.Format, msg.WordLen, msg.Rounds, msg.TurnTimeout, msg.PlayerName)

	case game.JoinRoomIntent:
		m.client.JoinRoom(msg.RoomId, msg.PlayerName)

	case game.RequestRematchIntent:
		m.client.RequestRematch()
		m.screenID = RequestRematchScreenID
		cmds = append(cmds, m.requestRematchScreen.Init())

	case game.DenyRematchIntent:
		if m.matchInfo.DeniedRematch {
			break
		}

		m.matchInfo.DeniedRematch = true
		m.matchInfo.RoomID = ""
		m.client.DenyRematch()
		cmds = append(cmds, emit(game.StartMenuIntent{}))

	case game.MakeGuessIntent:
		m.client.MakeGuess(msg.Guess)

	case game.ContinueIntent:
		logger.Debug("Continuing events, flushing buffer")
		m.eventsPaused = false
		m.flushing = true

		for _, bufferedEvent := range m.eventBuffer {
			eventCmd := m.handleEvent(bufferedEvent)
			if eventCmd != nil {
				cmds = append(cmds, eventCmd)
			}
		}

		m.flushing = false
		m.eventBuffer = nil

	case game.LeaveMatchIntent:
		m.client.LeaveMatch()
		cmds = append(cmds, emit(game.StartMenuIntent{}))

	case game.TypingIntent:
		if m.matchInfo.Mode == protocol.MULTI_REMOTE {
			m.client.Typing(msg.Value)
		}

	case game.ReadyNextRoundIntent:
		m.client.ReadyNextRound()
		if m.matchInfo.Mode == protocol.MULTI_REMOTE {
			m.game.state = game.StateWaitOpponentReady
		}

	case transport.ServerDisconnectedMsg:
		m.screenID = ServerDownScreenID

	case transport.EventMsg:
		cmds = append(cmds, transport.WaitForEvent(m.event))

		isHighPriority := m.isHighPriority(msg)

		if isHighPriority && m.eventsPaused {
			cmds = append(cmds, emit(game.ContinueIntent{}))
		}

		if m.eventsPaused && !isHighPriority {
			logger.Debug("UI paused, buffering event")
			m.eventBuffer = append(m.eventBuffer, msg)
		} else {
			eventCmd := m.handleEvent(msg)
			if eventCmd != nil {
				cmds = append(cmds, eventCmd)
			}
		}
		logger.Debug("[%s] event buffer: %v", m.matchInfo.PlayerName, m.eventBuffer)
	}

	switch m.screenID {

	case StartMenuScreenID:
		_, startMenuCmd := m.startMenuScreen.Update(msg)
		cmds = append(cmds, startMenuCmd)

	case GameConfigScreenID:
		updatedModel, formCmd := m.gameConfigMenu.Update(msg)
		m.gameConfigMenu = updatedModel.(*huh.Form)

		cmds = append(cmds, formCmd)

		if focusedField := m.gameConfigMenu.GetFocusedField(); focusedField != nil {
			if selectField, ok := focusedField.(*huh.Select[protocol.GameMode]); ok {
				newHovered, _ := selectField.Hovered()

				if newHovered != m.hoveredMode {
					m.hoveredMode = newHovered
					selectField.Description(GameModeDescriptions[m.hoveredMode])
					cmds = append(cmds, emit(forceRenderMsg{})) // trigger render of form with new description
				}
			}
		}

		if m.gameConfigMenu.State == huh.StateCompleted {
			if !*m.confirm {
				m.screenID = GameConfigScreenID
				m.gameConfigMenu, m.confirm = NewGameConfigMenu(m.matchInfo, &m.hoveredMode)

				return m, tea.Batch(tea.ClearScreen, formCmd)
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
			case protocol.MULTI_LOCAL:
				m.screenID = GameScreenID
				cmds = append(cmds, m.game.Init())
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

func (m *mainModel) View() string {
	var content string

	switch m.screenID {
	case StartMenuScreenID:
		content = lipgloss.NewStyle().Padding(1, 6).Render(m.startMenuScreen.View())
	case GameConfigScreenID:
		content = m.gameConfigMenu.View()
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

func (m *mainModel) handleEvent(eventMsg transport.EventMsg) tea.Cmd {
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

		m.matchInfo.Format = matchStartedEvent.Format
		m.matchInfo.WordLen = matchStartedEvent.WordLength
		m.matchInfo.TotalRounds = matchStartedEvent.Rounds
		m.matchInfo.RoundsPlayed = 0

		if m.matchInfo.Mode == protocol.MULTI_REMOTE {
			m.matchInfo.OpponentName = matchStartedEvent.OpponentName
		}

		m.matchInfo.RawTotalRounds = fmt.Sprintf("%d", matchStartedEvent.Rounds)
		m.matchInfo.RoundPoints = make([]int, matchStartedEvent.Rounds)

		m.game.input.CharLimit = m.matchInfo.WordLen
		m.game.input.Width = m.matchInfo.WordLen
		m.game.input.SetValue("")

		m.requestRematchScreen.opponentName = matchStartedEvent.OpponentName
		m.requestRematchScreen.opponentDeniedRematch = false

		if matchStartedEvent.TurnTimeout > 0 {
			m.game.turnTimeout = matchStartedEvent.TurnTimeout
		}

	case protocol.ROUND_STARTED:
		roundStartedEvent := &protocol.RoundStartedEvent{}
		if err := json.Unmarshal(msg, roundStartedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoundStartedEvent: %v", err)
			return nil
		}

		m.matchInfo.CurrentAttempt = 0
		m.matchInfo.MaxAttempts = roundStartedEvent.MaxAttempts
		m.matchInfo.CurrentRound = roundStartedEvent.RoundNumber

		m.game.initChallenges()
		m.game.input.SetValue("")

	case protocol.WAIT_GUESS:
		m.matchInfo.PlayerOnTurn = true
		m.game.state = game.StateWaitGuess
		m.game.input.Focus()
		m.game.input.SetValue("")
		m.game.err = nil

		var cmds []tea.Cmd
		cmds = append(cmds, m.game.setTurnTimer())
		if m.game.postRoundTimer.ID() != 0 {
			cmds = append(cmds, m.game.postRoundTimer.Stop())
		}
		return tea.Batch(cmds...)

	case protocol.WAIT_OPPONENT_GUESS:
		m.matchInfo.PlayerOnTurn = false
		m.game.state = game.StateWaitOpponentGuess
		m.game.input.SetValue("")
		m.game.err = nil

		if m.matchInfo.Mode == protocol.MULTI_LOCAL {
			m.game.input.Focus()
		} else {
			m.game.input.Blur()
		}

		var cmds []tea.Cmd
		cmds = append(cmds, m.game.setTurnTimer())
		if m.game.postRoundTimer.ID() != 0 {
			cmds = append(cmds, m.game.postRoundTimer.Stop())
		}
		return tea.Batch(cmds...)

	case protocol.WAIT_OPPONENT_JOIN:
		m.game.state = game.StateWaitOpponentJoin

	case protocol.GUESS_RESULT:
		guessResultEvent := &protocol.GuessResultEvent{}
		if err := json.Unmarshal(msg, guessResultEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling GuessResultEvent: %v", err)
			return nil
		}

		m.game.matchInfo.AddGuess(guessResultEvent.Guess)
		for i, feedback := range guessResultEvent.Feedback {
			challenge := m.game.challenges[i]
			if challenge.SolvedBy != protocol.OUTCOME_NONE {
				continue
			}

			challenge.Feedbacks[m.matchInfo.CurrentAttempt] = feedback
			if isSolved(feedback) {
				challenge.SolvedOnTurn = m.matchInfo.CurrentAttempt
				if m.matchInfo.PlayerOnTurn {
					challenge.SolvedBy = protocol.OUTCOME_PLAYER_WON
				} else {
					challenge.SolvedBy = protocol.OUTCOME_OPPONENT_WON
				}
			}
		}

		m.matchInfo.CurrentAttempt++

	case protocol.ROUND_FINISHED:
		roundFinishedEvent := &protocol.RoundFinishedEvent{}
		if err := json.Unmarshal(msg, roundFinishedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoundFinishedEvent: %v", err)
			return nil
		}

		m.game.state = game.StateRoundFinished
		m.matchInfo.CorrectWords = roundFinishedEvent.Words
		m.game.input.Blur()
		m.game.postRoundTimeout = roundFinishedEvent.PostRoundTimeout

		logger.Debug("Round points: %v", m.matchInfo.RoundPoints)
		m.matchInfo.RoundPoints[m.matchInfo.CurrentRound-1] = roundFinishedEvent.Points
		m.matchInfo.RoundsPlayed++

		logger.Debug("Pausing events...")
		if !m.flushing {
			m.eventsPaused = true
		}

		var cmds []tea.Cmd
		cmds = append(cmds, m.game.setPostRoundTimer())
		if m.game.turnTimer.ID() != 0 {
			cmds = append(cmds, m.game.turnTimer.Stop())
		}
		return tea.Batch(cmds...)

	case protocol.MATCH_FINISHED:
		matchFinishedEvent := &protocol.MatchFinishedEvent{}
		if err := json.Unmarshal(msg, matchFinishedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling MatchFinishedEvent: %v", err)
			return nil
		}

		m.game.state = game.StateMatchFinished

		return emit(MatchFinishedMsg{
			roundsPlayed: m.matchInfo.RoundsPlayed,
			roundPoints:  m.matchInfo.RoundPoints,
			matchOutcome: matchFinishedEvent.Outcome,
			opponentLeft: matchFinishedEvent.OpponentLeft,
		})

	case protocol.ROOM_CREATED:
		roomCreatedEvent := &protocol.RoomCreatedEvent{}
		if err := json.Unmarshal(msg, roomCreatedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoomCreatedEvent: %v", err)
			return nil
		}

		return emit(RoomCreatedMsg{
			roomID: roomCreatedEvent.RoomID,
		})

	case protocol.ROOM_JOIN_FAILED:
		roomJoinFailedEvent := &protocol.RoomJoinFailedEvent{}
		if err := json.Unmarshal(msg, roomJoinFailedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoomJoinFailedEvent: %v", err)
			return nil
		}

		m.screenID = GameConfigScreenID
		m.gameConfigMenu, m.confirm = NewGameConfigMenu(m.matchInfo, &m.hoveredMode)
		m.game.matchInfo.RoomValidationError = errors.New(roomJoinFailedEvent.Reason)

		// navigate to the RoomID input field
		m.gameConfigMenu.NextGroup()
		m.gameConfigMenu.NextGroup()
		m.gameConfigMenu.NextField()

		// simulate a key press to register the error message
		enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
		_, formCmd := m.gameConfigMenu.Update(enterMsg)
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

	case protocol.OPPONENT_LEFT:
		m.requestRematchScreen.opponentDeniedRematch = true

	default:
		logger.Info("Ignoring event: %s", event.Type)
	}

	return nil
}

func (m *mainModel) GetClient() *client.Client {
	return m.client
}

func (m *mainModel) SetSSHContext(ctx context.Context) {
	if ctx != nil {
		m.sshContext = ctx
	}
}

func (m *mainModel) isHighPriority(eventMsg transport.EventMsg) bool {
	msg := []byte(eventMsg)
	event := &protocol.EnvelopeEvent{}

	if err := json.Unmarshal(msg, event); err != nil {
		logger.Error("[handleEvent] error unmarshaling EnvelopeEvent: %s", err)
		return false
	}

	return slices.Contains(highPriorityEvents, event.Type)
}

// waitForContextDone is used to end the process immidiately
// after client disconnects from the SSH server
func waitForContextDone(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		<-ctx.Done()
		return tea.Quit()
	}
}

func isSolved(feedback []protocol.LetterFeedback) bool {
	for _, f := range feedback {
		if f != protocol.LETTER_CORRECT {
			return false
		}
	}
	return true
}
