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
		cmd = emit(game.JoinRoomIntent{
			RoomId:     m.matchInfo.RoomID,
			PlayerName: m.matchInfo.PlayerName,
		})
	} else {
		cmd = emit(game.CreateMatchIntent{
			Mode:       m.matchInfo.Mode,
			WordLen:    m.matchInfo.WordLen,
			Rounds:     m.matchInfo.TotalRounds,
			PlayerName: m.matchInfo.PlayerName,
		})
	}

	return tea.Batch(
		textinput.Blink,
		cmd,
	)
}

func (m *gameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

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
		var inputCmd tea.Cmd
		m.input, inputCmd = m.input.Update(msg)
		cmds = append(cmds, inputCmd)

		if m.state == game.StateWaitGuess {
			cmds = append(cmds, emit(game.TypingIntent{Value: m.input.Value()}))
		}
		return m, tea.Batch(cmds...)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *gameModel) View() string {
	// 1. Render and measure the grid
	guessGrid := ui.ViewGuessGrid(
		m.guesses,
		m.input.Value(),
		m.matchInfo.MaxAttempts,
		m.matchInfo.WordLen,
		m.input.Focused(),
	)

	gridWidth := lipgloss.Width(guessGrid)

	// 2. Define the "Game Area"
	// Instead of m.width, we take the grid and add a fixed amount of padding (e.g., 12 chars).
	// This makes the UI look consistent regardless of how big the terminal window is.

	// 3. Components
	player1 := fmt.Sprintf("%s%s", ui.OutcomeBlockStyle(ui.Purple).Render(""), m.matchInfo.PlayerName)
	player2 := fmt.Sprintf("%s%s", ui.OutcomeBlockStyle(ui.Red).Render(""), m.matchInfo.OpponentName)
	outcomes := ui.ViewRoundOutcomes(m.matchInfo.RoundOutcomes)

	gameAreaWidth := gridWidth + lipgloss.Width(player1) + lipgloss.Width(player2)

	// 4. Calculate gaps based on gameAreaWidth
	totalComponentsWidth := lipgloss.Width(player1) + lipgloss.Width(outcomes) + lipgloss.Width(player2)
	totalSpace := max(0, gameAreaWidth-totalComponentsWidth)

	gapWidth := totalSpace / 2
	leftSpacer := strings.Repeat(" ", gapWidth)
	rightSpacer := strings.Repeat(" ", totalSpace-gapWidth) // Handles odd numbers

	// 5. Build the header
	headerRow := lipgloss.JoinHorizontal(
		lipgloss.Center,
		player1,
		leftSpacer,
		outcomes,
		rightSpacer,
		player2,
	)

	// 6. Stack them. JoinVertical(lipgloss.Center) ensures the
	// narrower guessGrid is centered under the slightly wider headerRow.
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		headerRow,
		"\n",
		guessGrid,
	)

	// 7. Center the "Game Area" on the actual screen
	// This is the only place we use m.width.
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

func emit(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
