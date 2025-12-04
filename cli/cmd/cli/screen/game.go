package screen

import (
	"encoding/json"
	"guessh/cmd/cli/client"
	"guessh/cmd/cli/protocol"
	"guessh/cmd/cli/transport"
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
	StateWaitingGuess
	StateWaitingGuessResult
	StateWaitOpponentJoin
	StateWaitOpponentGuess
	StateRoundFinished
	StateMatchFinished
)

type model struct {
	client  *client.Client
	input   textinput.Model
	guesses []*protocol.Guess
	state   GameState
	msg     chan transport.EventMsg
	err     error
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
		client: c,
		input:  ti,
		state:  StateInit,
		msg:    make(chan transport.EventMsg),
		err:    nil,
	}
}

func (m model) Init() tea.Cmd {
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
			m.client.MakeGuess(m.input.Value())
			m.input.SetValue("")
		}
	case transport.EventMsg:
		log.Printf("New event received: %s\n", string(msg))
		m.handleEvent(msg)
		return m, transport.WaitForEvent(m.msg)
	case error:
		m.err = msg
		return m, nil
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.input.View()
}

func (m model) handleEvent(eventMsg transport.EventMsg) {
	msg := []byte(strings.TrimRight(string(eventMsg), " \t\n\r\x00"))
	event := &protocol.EnvelopeMessage{}

	if err := json.Unmarshal([]byte(msg), event); err != nil {
		log.Printf("[handleEvent] error unmarshaling EnvelopeMessage: %s", err)
		return
	}

	switch event.Type {

	case protocol.MATCH_STARTED:
		m.state = StateMatchStarted

	case protocol.ROUND_STARTED:
		m.state = StateMatchStarted

	case protocol.GUESS_RESULT:
		// TODO: If MULTI_REMOTE, toggle StateWaitingGuess and StateWaitOpponentGuess
		m.state = StateWaitingGuess

		guessResultEvent := &protocol.GuessResultMessage{}

		if err := json.Unmarshal(msg, guessResultEvent); err != nil {
			log.Printf("[handleEvent] error unmarshaling GuessResultMessage: %v", err)
			return
		}

		m.guesses = append(m.guesses, protocol.NewGuess(guessResultEvent.Guess, guessResultEvent.Feedback))

	case protocol.ROUND_FINISHED:
		m.state = StateRoundFinished

	case protocol.MATCH_FINISHED:
		m.state = StateMatchFinished

	}

	log.Printf("[handleEvent] Event type: %s", event.Type)
}
