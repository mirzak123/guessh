package screen

import (
	"encoding/json"
	"fmt"
	"guessh/internal/game"
	"guessh/internal/logger"
	"guessh/internal/protocol"
	"guessh/internal/transport"
	"guessh/internal/ui"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type gameModel struct {
	width, height int
	matchInfo     *game.MatchInfo
	input         textinput.Model
	guesses       []*protocol.Guess
	state         game.GameState
	roundInfo     *game.RoundInfo
}

func NewGame(matchInfo *game.MatchInfo) *gameModel {
	logger.Debug("Calling NewGame")
	ti := textinput.New()
	ti.CharLimit = matchInfo.WordLen
	ti.Width = matchInfo.WordLen

	return &gameModel{
		matchInfo: matchInfo,
		input:     ti,
		state:     game.StateInit,
		roundInfo: game.NewRoundInfo(),
	}
}

func (m *gameModel) Init() tea.Cmd {
	logger.Info("[Game Init] matchInfo.Mode: %s", m.matchInfo.Mode)
	var err error

	if m.matchInfo.TotalRounds, err = strconv.Atoi(m.matchInfo.RawTotalRounds); err != nil {
		logger.Error("[Client.CreateMatch] Failed to convert matchInfo.RawTotalRounds after it passed validation: %v", err)
		os.Exit(1)
	}

	return tea.Batch(
		textinput.Blink,
		emit(game.CreateMatchIntent{
			Mode:    m.matchInfo.Mode,
			WordLen: m.matchInfo.WordLen,
			Rounds:  m.matchInfo.TotalRounds,
		}),
	)
}

func (m *gameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logger.Debug("Game State %d", m.state)
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		logger.Debug("[Update] Window resizing...")
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			switch m.state {
			case game.StateRoundFinished:
				return m, emit(game.ContinueIntent{})
			case game.StateWaitGuess:
				v := m.input.Value()

				if len(v) == m.matchInfo.WordLen { // do nothing if not enough letters
					m.input.SetValue("")
					return m, emit(game.MakeGuessIntent{Guess: m.input.Value()})
				}
			}

		}

	case transport.EventMsg:
		logger.Info("New event received: %s\n", string(msg))

		msgFromEvent := m.handleEvent(msg)
		if msgFromEvent != nil {
			return m, emit(msgFromEvent)
		}

	case error:
		return m, nil
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *gameModel) View() string {
	var body, header, footer string

	guessGrid := ui.ViewGuessGrid(m.guesses, m.input.Value(), m.matchInfo.MaxAttempts, m.matchInfo.WordLen)
	continueMsg := "Press enter to continue..."

	header += fmt.Sprintf(
		"Round: %d/%s\tGuessed correctly: %d\n",
		m.matchInfo.CurrentRound,
		m.matchInfo.RawTotalRounds,
		m.matchInfo.RoundsWon,
	)

	centeredGrid := lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Center,
		guessGrid,
	)

	body = centeredGrid

	if m.state == game.StateRoundFinished {
		if m.roundInfo.Success {
			footer = continueMsg
		} else {
			footer = fmt.Sprintf("Correct word: %s\n%s", m.roundInfo.Word, continueMsg)
		}
	}

	view := strings.Join([]string{header, body, footer}, "\n")

	return lipgloss.NewStyle().
		Width(m.width).
		AlignHorizontal(lipgloss.Center).
		Render(view)
}

func (m *gameModel) handleEvent(eventMsg transport.EventMsg) tea.Msg {
	msg := []byte(eventMsg)
	event := &protocol.EnvelopeMessage{}

	if err := json.Unmarshal(msg, event); err != nil {
		logger.Error("[handleEvent] error unmarshaling EnvelopeMessage: %s", err)
		return nil
	}

	logger.Info("[handleEvent] Event type: %s", event.Type)

	switch event.Type {

	case protocol.MATCH_STARTED:
		logger.Debug("Changing state to StateMatchStarted")
		m.state = game.StateMatchStarted

	case protocol.ROUND_STARTED:
		m.matchInfo.CurrentRound++
		m.roundInfo = game.NewRoundInfo()
		m.input.Focus()

		roundStartedEvent := &protocol.RoundStartedMessage{}

		if err := json.Unmarshal(msg, roundStartedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoundStartedMessage: %v", err)
			return nil
		}
		m.matchInfo.MaxAttempts = roundStartedEvent.MaxAttempts
		m.guesses = nil

	case protocol.WAIT_GUESS:
		m.state = game.StateWaitGuess

	case protocol.WAIT_OPPONENT_GUESS:
		m.state = game.StateWaitOpponentGuess

	case protocol.WAIT_OPPONENT_JOIN:
		m.state = game.StateWaitOpponentJoin

	case protocol.GUESS_RESULT:
		guessResultEvent := &protocol.GuessResultMessage{}

		if err := json.Unmarshal(msg, guessResultEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling GuessResultMessage: %v", err)
			return nil
		}

		m.guesses = append(m.guesses, protocol.NewGuess(guessResultEvent.Guess, guessResultEvent.Feedback))

	case protocol.ROUND_FINISHED:
		roundFinishedEvent := &protocol.RoundFinishedMessage{}

		if err := json.Unmarshal(msg, roundFinishedEvent); err != nil {
			logger.Error("[handleEvent] error unmarshaling RoundFinishedMessage: %v", err)
			return nil
		}

		m.state = game.StateRoundFinished
		m.roundInfo.Word = roundFinishedEvent.Word
		m.roundInfo.Success = roundFinishedEvent.Success

		m.input.Blur()

		if roundFinishedEvent.Success {
			m.matchInfo.RoundsWon++
		}

		return game.PauseIntent{}

	case protocol.MATCH_FINISHED:
		m.state = game.StateMatchFinished
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

	}

	return nil
}

func emit(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
