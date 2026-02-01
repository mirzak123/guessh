package screen

import (
	"encoding/json"
	"fmt"
	"guessh/internal/client"
	"guessh/internal/protocol"
	"guessh/internal/transport"
	"guessh/internal/ui"
	"log"
	"net"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type GameState int

const (
	StateInit GameState = iota
	StateMatchStarted
	StateWaitGuess
	StateWaitingGuessResult
	StateWaitOpponentJoin
	StateWaitOpponentGuess
	StateRoundFinished
	StateMatchFinished
)

type GameFinishedMsg struct{}

type model struct {
	client    *client.Client
	matchInfo *protocol.MatchInfo
	input     textinput.Model
	guesses   []*protocol.Guess
	state     GameState
	msg       chan transport.EventMsg
	err       error
	uiPaused  bool
}

func NewGame(matchInfo *protocol.MatchInfo) model {
	conn, err := net.Dial("tcp", "localhost:2480")
	if err != nil {
		log.Fatalf("net.Dial error: %v", err)
	}

	c := client.NewClient(conn)

	ti := textinput.New()
	ti.CharLimit = matchInfo.WordLen
	ti.Width = matchInfo.WordLen
	ti.Focus()

	return model{
		client:    c,
		matchInfo: matchInfo,
		input:     ti,
		state:     StateInit,
		msg:       make(chan transport.EventMsg),
		err:       nil,
		uiPaused:  false,
	}
}

func (m model) Init() tea.Cmd {
	log.Printf("[Game Init] matchInfo.Mode: %s", m.matchInfo.Mode)
	m.client.CreateMatch(m.matchInfo)

	return tea.Batch(
		textinput.Blink,
		transport.ListenForActivity(m.client.Conn, m.msg),
		transport.WaitForEvent(m.msg))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.uiPaused {
				// Resume processing events
				m.uiPaused = false
				return m, transport.WaitForEvent(m.msg)
			}

			v := m.input.Value()

			if len(v) == m.matchInfo.WordLen { // do nothing if not enough letters
				m.client.MakeGuess(m.input.Value())
				m.input.SetValue("")
			}
		}

	case transport.EventMsg:
		log.Printf("State: %d", m.state)
		log.Printf("New event received: %s\n", string(msg))

		m, msgFromEvent := m.handleEvent(msg)
		if msgFromEvent != nil {
			return m, func() tea.Msg { return msgFromEvent } // TODO: This is ugly, fix this
		}

		if m.uiPaused {
			// Pause processing events from channel
			return m, nil
		}
		return m, transport.WaitForEvent(m.msg)

	case error:
		m.err = msg
		return m, nil
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var view, header string

	guessGrid := ui.ViewGuessGrid(m.guesses, m.input.Value(), m.matchInfo.MaxAttempts, m.matchInfo.WordLen)
	continueMsg := "Press enter to continue..."

	if m.uiPaused {
		view = fmt.Sprintf("%s\n%s", guessGrid, continueMsg)
	} else {
		view = fmt.Sprintf("%s\n\n", guessGrid)
	}

	header += fmt.Sprintf("Round: %d/%s\n", m.matchInfo.CurrentRound, m.matchInfo.RawTotalRounds)
	view = header + view

	return lipgloss.NewStyle().Width(len(continueMsg)).Render(view)
}

func (m model) handleEvent(eventMsg transport.EventMsg) (model, tea.Msg) {
	msg := []byte(eventMsg)
	event := &protocol.EnvelopeMessage{}

	if err := json.Unmarshal(msg, event); err != nil {
		log.Printf("[handleEvent] error unmarshaling EnvelopeMessage: %s", err)
		return m, nil
	}

	switch event.Type {

	case protocol.MATCH_STARTED:
		m.state = StateMatchStarted

	case protocol.ROUND_STARTED:
		m.matchInfo.CurrentRound++
		roundStartedEvent := &protocol.RoundStartedMessage{}

		if err := json.Unmarshal(msg, roundStartedEvent); err != nil {
			log.Printf("[handleEvent] error unmarshaling RoundStartedMessage: %v", err)
			return m, nil
		}
		m.state = StateMatchStarted
		m.matchInfo.MaxAttempts = roundStartedEvent.MaxAttempts
		m.guesses = nil

	case protocol.WAIT_GUESS:
		m.state = StateWaitGuess

	case protocol.WAIT_OPPONENT_GUESS:
		m.state = StateWaitOpponentGuess

	case protocol.WAIT_OPPONENT_JOIN:
		m.state = StateWaitOpponentJoin

	case protocol.GUESS_RESULT:
		guessResultEvent := &protocol.GuessResultMessage{}

		if err := json.Unmarshal(msg, guessResultEvent); err != nil {
			log.Printf("[handleEvent] error unmarshaling GuessResultMessage: %v", err)
			return m, nil
		}

		m.guesses = append(m.guesses, protocol.NewGuess(guessResultEvent.Guess, guessResultEvent.Feedback))

	case protocol.ROUND_FINISHED:
		m.state = StateRoundFinished
		m.uiPaused = true

	case protocol.MATCH_FINISHED:
		m.state = StateMatchFinished
		m.uiPaused = true
		return m, GameFinishedMsg{}
	}

	log.Printf("[handleEvent] Event type: %s", event.Type)
	return m, nil
}
