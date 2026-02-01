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
	"strconv"
	"strings"

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

type MatchInfo struct {
	mode           protocol.GameMode
	wordLen        int
	currentRound   int
	rawTotalRounds string
	totalRounds    int
	maxAttempts    int
	roundsGuessed  int
}

func NewMatchInfo() *MatchInfo {
	return &MatchInfo{
		currentRound: 0,
	}
}

type RoundInfo struct {
	word    string
	success bool
}

func NewRoundInfo() *RoundInfo {
	return &RoundInfo{}
}

type gameModel struct {
	width, height int
	client        *client.Client
	matchInfo     *MatchInfo
	input         textinput.Model
	guesses       []*protocol.Guess
	state         GameState
	msg           chan transport.EventMsg
	err           error
	roundInfo     *RoundInfo
	uiPaused      bool
}

func NewGame(matchInfo *MatchInfo) gameModel {
	conn, err := net.Dial("tcp", "localhost:2480")
	if err != nil {
		log.Fatalf("net.Dial error: %v", err)
	}

	c := client.NewClient(conn)

	ti := textinput.New()
	ti.CharLimit = matchInfo.wordLen
	ti.Width = matchInfo.wordLen

	return gameModel{
		client:    c,
		matchInfo: matchInfo,
		input:     ti,
		state:     StateInit,
		msg:       make(chan transport.EventMsg),
		err:       nil,
		roundInfo: NewRoundInfo(),
	}
}

func (m gameModel) Init() tea.Cmd {
	log.Printf("[Game Init] matchInfo.Mode: %s", m.matchInfo.mode)
	var err error

	if m.matchInfo.totalRounds, err = strconv.Atoi(m.matchInfo.rawTotalRounds); err != nil {
		log.Fatalf("[Client.CreateMatch] Failed to convert matchInfo.RawTotalRounds after it passed validation: %v", err)
	}

	m.client.CreateMatch(m.matchInfo.mode, m.matchInfo.wordLen, m.matchInfo.totalRounds)

	return tea.Batch(
		textinput.Blink,
		transport.ListenForActivity(m.client.Conn, m.msg),
		transport.WaitForEvent(m.msg))
}

func (m gameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		log.Print("[Update] Window resizing...")
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

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

			if len(v) == m.matchInfo.wordLen { // do nothing if not enough letters
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

func (m gameModel) View() string {
	var body, header, footer string

	guessGrid := ui.ViewGuessGrid(m.guesses, m.input.Value(), m.matchInfo.maxAttempts, m.matchInfo.wordLen)
	continueMsg := "Press enter to continue..."

	header += fmt.Sprintf(
		"Round: %d/%s\tGuessed correctly: %d\n",
		m.matchInfo.currentRound,
		m.matchInfo.rawTotalRounds,
		m.matchInfo.roundsGuessed,
	)

	centeredGrid := lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Center,
		guessGrid,
	)

	body = centeredGrid

	if m.uiPaused {
		if m.roundInfo.success {
			footer = continueMsg
		} else {
			footer = fmt.Sprintf("Correct word: %s\n%s", m.roundInfo.word, continueMsg)
		}
	}

	view := strings.Join([]string{header, body, footer}, "\n")

	return lipgloss.NewStyle().
		Width(m.width).
		AlignHorizontal(lipgloss.Center).
		Render(view)
}

func (m gameModel) handleEvent(eventMsg transport.EventMsg) (gameModel, tea.Msg) {
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
		m.matchInfo.currentRound++
		m.roundInfo = NewRoundInfo()
		m.input.Focus()

		roundStartedEvent := &protocol.RoundStartedMessage{}

		if err := json.Unmarshal(msg, roundStartedEvent); err != nil {
			log.Printf("[handleEvent] error unmarshaling RoundStartedMessage: %v", err)
			return m, nil
		}
		m.state = StateMatchStarted
		m.matchInfo.maxAttempts = roundStartedEvent.MaxAttempts
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
		roundFinishedEvent := &protocol.RoundFinishedMessage{}
		log.Print(roundFinishedEvent)

		if err := json.Unmarshal(msg, roundFinishedEvent); err != nil {
			log.Printf("[handleEvent] error unmarshaling RoundFinishedMessage: %v", err)
			return m, nil
		}

		m.state = StateRoundFinished
		m.uiPaused = true
		m.roundInfo.word = roundFinishedEvent.Word
		m.roundInfo.success = roundFinishedEvent.Success

		m.input.Blur()

		if roundFinishedEvent.Success {
			m.matchInfo.roundsGuessed++
		}

	case protocol.MATCH_FINISHED:
		m.state = StateMatchFinished
		m.uiPaused = true
		return m, MatchFinishedMsg{roundsPlayed: m.matchInfo.totalRounds}
	}

	log.Printf("[handleEvent] Event type: %s", event.Type)
	return m, nil
}
