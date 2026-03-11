package screen

import (
	"fmt"
	"guessh/internal/game"
	"guessh/internal/logger"
	"guessh/internal/protocol"
	"guessh/internal/ui"
	"os"
	"slices"
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
	logger.Debug("Creating new game model")
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
			Format:     m.matchInfo.Format,
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

	shouldRegisterInput := false

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
		case tea.KeyBackspace:
			shouldRegisterInput = true
		case tea.KeyRunes:
			r := &msg.Runes[0]
			if *r >= 'A' && *r <= 'Z' {
				msg.Runes[0] = *r + ('a' - 'A')
			}
			if *r < 'a' || *r > 'z' {
				return m, nil
			}
			shouldRegisterInput = true
		case tea.KeyEnter:
			logger.Debug("Game state: [%s]", m.state)
			if m.state == game.StateRoundFinished {
				return m, emit(game.ContinueIntent{})
			}
			if m.state == game.StateWaitOpponentGuess && m.matchInfo.Mode != protocol.MULTI_LOCAL {
				break
			}
			if m.state == game.StateWaitGuess || m.state == game.StateWaitOpponentGuess {
				v := m.input.Value()

				if len(v) == m.matchInfo.WordLen && !m.alreadyGuessed(v) {
					if m.validateGuess(v) {
						m.input.SetValue("")
						return m, emit(game.MakeGuessIntent{Guess: v})
					}
				}
			}
		}

		if shouldRegisterInput {
			var inputCmd tea.Cmd
			m.input, inputCmd = m.input.Update(msg)
			cmds = append(cmds, inputCmd)

			if m.state == game.StateWaitGuess {
				cmds = append(cmds, emit(game.TypingIntent{Value: m.input.Value()}))
			}
		}
		return m, tea.Batch(cmds...)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *gameModel) View() string {
	var wordNum int
	switch m.matchInfo.Format {
	case protocol.WORDLE:
		wordNum = 1
	case protocol.QUORDLE:
		wordNum = 4
	}

	gridStyle := lipgloss.NewStyle().MarginRight(4)

	var guessGrids []string
	for i := 0; i < wordNum; i++ {
		grid := ui.ViewGuessGrid(
			m.guesses,
			i,
			m.input.Value(),
			m.matchInfo.MaxAttempts,
			m.matchInfo.WordLen,
			m.state,
		)

		// Apply the margin to all but the last element
		if i < wordNum-1 {
			grid = gridStyle.Render(grid)
		}

		guessGrids = append(guessGrids, grid)
	}

	gridWidth := lipgloss.Width(guessGrids[0])

	var (
		p1Symbol = ui.PlayerBlock()
		p2Symbol string
		p1Name   string
		p2Name   string
	)

	if m.matchInfo.Mode == protocol.SINGLE {
		p2Symbol = ui.NoOpponentBlock()
		p1Name = "You"
		p2Name = "No Opponent"
	} else {
		p2Symbol = ui.OpponentBlock()
		p1Name = m.matchInfo.PlayerName
		p2Name = m.matchInfo.OpponentName
	}

	player1 := fmt.Sprintf("%s%s",
		lipgloss.NewStyle().MarginRight(1).Render(p1Symbol),
		p1Name,
	)
	player2 := fmt.Sprintf("%s%s",
		lipgloss.NewStyle().MarginRight(1).Render(p2Name),
		p2Symbol,
	)

	p1w := lipgloss.Width(player1)
	p2w := lipgloss.Width(player2)
	maxPlayerWidth := max(p1w, p2w)

	if p1w < maxPlayerWidth {
		player1 += strings.Repeat(" ", maxPlayerWidth-p1w)
	}
	if p2w < maxPlayerWidth {
		player2 = strings.Repeat(" ", maxPlayerWidth-p2w) + player2
	}

	outcomes := ui.ViewRoundOutcomes(m.matchInfo.RoundPoints, m.matchInfo.Format, m.matchInfo.RoundsPlayed)

	gameAreaWidth := gridWidth + maxPlayerWidth*2

	totalComponentsWidth :=
		maxPlayerWidth +
			lipgloss.Width(outcomes) +
			maxPlayerWidth

	totalSpace := max(0, gameAreaWidth-totalComponentsWidth)

	gapWidth := totalSpace / 2
	leftSpacer := strings.Repeat(" ", gapWidth)
	rightSpacer := strings.Repeat(" ", totalSpace-gapWidth)

	headerRow := lipgloss.JoinHorizontal(
		lipgloss.Center,
		player1,
		leftSpacer,
		outcomes,
		rightSpacer,
		player2,
	)

	emptyLine := lipgloss.NewStyle().
		Height(1).
		Render("")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		headerRow,
		emptyLine,
		emptyLine,
		lipgloss.JoinHorizontal(lipgloss.Center, guessGrids...),
		emptyLine,
		m.statusBar(),
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

func (m *gameModel) statusBar() string {
	var content string

	if m.state == game.StateRoundFinished {
		var line1 string

		points := m.matchInfo.RoundPoints[m.matchInfo.CurrentRound-1]
		if points > 0 {
			if m.matchInfo.Mode == protocol.MULTI_LOCAL {
				line1 = fmt.Sprintf("%s Round Won by %s",
					ui.PlayerBlock(),
					ui.PurpleText.Render(m.matchInfo.PlayerName))
			} else {
				line1 = fmt.Sprintf("%s Round Won", ui.OpponentBlock())
			}
		} else if points < 0 {
			if m.matchInfo.Mode == protocol.MULTI_LOCAL {
				line1 = fmt.Sprintf("%s Round Won by %s",
					ui.OpponentBlock(),
					ui.RoseText.Render(m.matchInfo.OpponentName))
			} else {
				line1 = fmt.Sprintf("%s Round Lost", ui.OpponentBlock())
			}
		} else {
			line1 = fmt.Sprintf(
				"%s Not Guessed - Correct word: %s",
				ui.DrawBlock(),
				m.roundInfo.Word)
		}

		line2 := ui.GrayText.Render("Press Enter to continue")

		content = lipgloss.JoinVertical(
			lipgloss.Center,
			line1,
			line2,
		)

	} else {
		content = ui.SmallLogo()
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Height(2).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Top).
		Render(content)
}

func (m *gameModel) validateGuess(guess string) bool {
	switch len(guess) {
	case 5:
		return slices.Index(game.FiveLetterWords, guess) != -1
	case 6:
		return slices.Index(game.SixLetterWords, guess) != -1
	case 7:
		return slices.Index(game.SevenLetterWords, guess) != -1
	default:
		logger.Error("[validateGuess] received value of length %d", len(guess))
		return false
	}
}

func emit(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

func (m *gameModel) alreadyGuessed(word string) bool {
	for _, guess := range m.guesses {
		if guess.Word == word {
			return true
		}
	}
	return false
}
