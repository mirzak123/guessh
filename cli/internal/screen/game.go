package screen

import (
	"encoding/json"
	"fmt"
	"guessh/internal/client"
	"guessh/internal/protocol"
	"guessh/internal/transport"
	"log"
	"net"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	State     GameState
	msg       chan transport.EventMsg
	err       error
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
		State:     StateInit,
		msg:       make(chan transport.EventMsg),
		err:       nil,
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
			v := m.input.Value()

			if len(v) == m.matchInfo.WordLen { // do nothing if not enough letters
				m.client.MakeGuess(m.input.Value())
				m.input.SetValue("")
			}
		}

	case transport.EventMsg:
		log.Printf("State: %d", m.State)
		log.Printf("New event received: %s\n", string(msg))
		m, msgFromEvent := m.handleEvent(msg)
		if msgFromEvent != nil {
			return m, func() tea.Msg { return msgFromEvent } // TODO: This is ugly, fix this
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
	guessRows := make([]string, len(m.guesses))

	for i, guess := range m.guesses {
		guessRows[i] = guess.View()
	}

	return fmt.Sprintf("%s\n%s", m.input.View(), strings.Join(guessRows, "\n"))
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
		m.State = StateMatchStarted

	case protocol.ROUND_STARTED:
		m.State = StateMatchStarted

	case protocol.WAIT_GUESS:
		m.State = StateWaitGuess

	case protocol.WAIT_OPPONENT_GUESS:
		m.State = StateWaitOpponentGuess

	case protocol.WAIT_OPPONENT_JOIN:
		m.State = StateWaitOpponentJoin

	case protocol.GUESS_RESULT:
		guessResultEvent := &protocol.GuessResultMessage{}

		if err := json.Unmarshal(msg, guessResultEvent); err != nil {
			log.Printf("[handleEvent] error unmarshaling GuessResultMessage: %v", err)
			return m, nil
		}

		m.guesses = append(m.guesses, protocol.NewGuess(guessResultEvent.Guess, guessResultEvent.Feedback))

	case protocol.ROUND_FINISHED:
		m.State = StateRoundFinished
		m.guesses = nil

	case protocol.MATCH_FINISHED:
		m.State = StateMatchFinished
		return m, GameFinishedMsg{}
	}

	log.Printf("[handleEvent] Event type: %s", event.Type)
	return m, nil
}
