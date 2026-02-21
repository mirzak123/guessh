package screen

import (
	"fmt"
	"guessh/internal/game"
	"guessh/internal/logger"
	"guessh/internal/protocol"
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
	ti.Blur()

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

	if !m.matchInfo.JoinExisting {
		if m.matchInfo.TotalRounds, err = strconv.Atoi(m.matchInfo.RawTotalRounds); err != nil {
			logger.Error("[Client.CreateMatch] Failed to convert matchInfo.RawTotalRounds after it passed validation: %v", err)
			os.Exit(1)
		}
	}

	var cmd tea.Cmd

	if m.matchInfo.JoinExisting {
		cmd = emit(game.JoinRoom{RoomId: m.matchInfo.RoomID})
	} else {
		cmd = emit(game.CreateMatchIntent{
			Mode:    m.matchInfo.Mode,
			WordLen: m.matchInfo.WordLen,
			Rounds:  m.matchInfo.TotalRounds,
		})
	}

	return tea.Batch(
		textinput.Blink,
		cmd,
	)
}

func (m *gameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			logger.Debug("Game state: [%s]", m.state)
			switch m.state {
			case game.StateRoundFinished:
				return m, emit(game.ContinueIntent{})
			case game.StateWaitGuess:
				v := m.input.Value()

				if len(v) == m.matchInfo.WordLen { // do nothing if not enough letters
					m.input.SetValue("")
					return m, emit(game.MakeGuessIntent{Guess: v})
				}
			}
		}

	case error:
		return m, nil
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *gameModel) View() string {
	var body, header, footer string

	guessGrid := ui.ViewGuessGrid(
		m.guesses,
		m.input.Value(),
		m.matchInfo.MaxAttempts,
		m.matchInfo.WordLen,
		m.input.Focused(),
	)
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

func emit(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
