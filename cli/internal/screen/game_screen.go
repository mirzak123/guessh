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
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type gameModel struct {
	width, height    int
	matchInfo        *game.MatchInfo
	input            textinput.Model
	state            game.GameState
	roundInfo        *game.RoundInfo
	guesses          []string
	challenges       []*protocol.WordChallenge
	challengesLen    int
	turnTimer        timer.Model
	postRoundTimer   timer.Model
	turnTimeout      int
	postRoundTimeout int
	err              error
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

	if m.matchInfo.RawTurnTimeout != "" {
		if m.matchInfo.TurnTimeout, err = strconv.Atoi(m.matchInfo.RawTurnTimeout); err != nil {
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
			Mode:        m.matchInfo.Mode,
			Format:      m.matchInfo.Format,
			WordLen:     m.matchInfo.WordLen,
			Rounds:      m.matchInfo.TotalRounds,
			TurnTimeout: m.matchInfo.TurnTimeout,
			PlayerName:  m.matchInfo.PlayerName,
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
			m.err = nil
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
				if !m.isLastRound() {
					return m, emit(game.ReadyNextRoundIntent{})
				} else {
					return m, emit(game.ContinueIntent{})
				}
			}
			if m.state == game.StateWaitOpponentGuess && m.matchInfo.Mode != protocol.MULTI_LOCAL {
				break
			}
			if m.state == game.StateRoundFinished && m.matchInfo.Mode == protocol.MULTI_REMOTE {
				m.state = game.StateWaitOpponentReady
				break
			}
			if m.state == game.StateWaitGuess || m.state == game.StateWaitOpponentGuess {
				v := m.input.Value()

				if len(v) == m.matchInfo.WordLen {
					if ok, err := m.validateGuess(v); ok {
						m.input.SetValue("")
						return m, emit(game.MakeGuessIntent{Guess: v})
					} else {
						m.err = err
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

	var inputCmd, timerCmd, postTurnTimerCmd tea.Cmd

	m.input, inputCmd = m.input.Update(msg)
	m.turnTimer, timerCmd = m.turnTimer.Update(msg)
	m.postRoundTimer, postTurnTimerCmd = m.postRoundTimer.Update(msg)

	cmds = append(cmds, inputCmd, timerCmd, postTurnTimerCmd)

	return m, tea.Batch(cmds...)
}

func (m *gameModel) View() string {
	if m.challengesLen == 0 {
		return ""
	}

	gridStyle := lipgloss.NewStyle().MarginRight(6)

	var guessGrids []string
	for i := range m.challengesLen {
		grid := ui.ViewGuessGrid(
			m.guesses,
			m.challenges[i],
			m.input.Value(),
			m.matchInfo.CurrentAttempt,
			m.matchInfo.MaxAttempts,
			m.matchInfo.WordLen,
			m.state,
		)

		// Apply the margin to all but the last element
		if i < m.challengesLen-1 {
			grid = gridStyle.Render(grid)
		}

		guessGrids = append(guessGrids, grid)
	}

	gridView := lipgloss.JoinHorizontal(lipgloss.Center, guessGrids...)
	gridWidth := lipgloss.Width(gridView)

	var (
		p1Symbol  = ui.PlayerBlock()
		p2Symbol  string
		p1Name    string
		p2Name    string
		countdown string
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

	if m.turnTimer.Running() {
		seconds := int(m.turnTimer.Timeout.Seconds())
		countdown = fmt.Sprintf("%3d", seconds)

		var style lipgloss.Style
		if seconds < 10 {
			style = ui.RoseText
		} else if seconds < 20 {
			style = ui.YellowText
		} else {
			style = ui.GreenText
		}
		countdown = style.Render(countdown)
	} else {
		countdown = strings.Repeat(" ", 3)
	}

	cdW := lipgloss.Width(countdown)
	if m.matchInfo.PlayerOnTurn {
		player1 = fmt.Sprintf("%s %s", player1, countdown)
		player2 = fmt.Sprintf("%s %s", strings.Repeat(" ", cdW), player2)
	} else {
		player1 = fmt.Sprintf("%s %s", player1, strings.Repeat(" ", cdW))
		player2 = fmt.Sprintf("%s %s", countdown, player2)
	}

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

	gameAreaWidth := gridWidth + maxPlayerWidth*2 // TODO: verify this is correct

	totalComponentsWidth := maxPlayerWidth + lipgloss.Width(outcomes) + maxPlayerWidth

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
		gridView,
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
	var line1, line2, line3 string

	if m.state == game.StateRoundFinished || m.state == game.StateWaitOpponentReady {
		var outcome string

		points := m.matchInfo.RoundPoints[m.matchInfo.CurrentRound-1]
		if points > 0 {
			if m.matchInfo.Mode == protocol.MULTI_LOCAL {
				outcome = fmt.Sprintf("%s Round won by %s",
					ui.PlayerBlock(),
					ui.PurpleText.Render(m.matchInfo.PlayerName))
			} else {
				outcome = fmt.Sprintf("%s Round won", ui.PlayerBlock())
			}
		} else if points < 0 {
			if m.matchInfo.Mode == protocol.MULTI_LOCAL {
				outcome = fmt.Sprintf("%s Round won by %s",
					ui.OpponentBlock(),
					ui.RoseText.Render(m.matchInfo.OpponentName))
			} else {
				outcome = fmt.Sprintf("%s Round lost", ui.OpponentBlock())
			}
		}
		words := []string{}

		for i, challenge := range m.challenges {
			word := m.matchInfo.CorrectWords[i]
			switch challenge.SolvedBy {
			case protocol.OUTCOME_NONE:
				words = append(words, ui.GrayText.Render(word))
			case protocol.OUTCOME_PLAYER_WON:
				words = append(words, ui.PurpleText.Render(word))
			case protocol.OUTCOME_OPPONENT_WON:
				words = append(words, ui.RoseText.Render(word))
			}
		}

		line1 = outcome
		switch m.state {
		case game.StateRoundFinished:
			line2 = ui.GrayText.Render("Press Enter to continue")
		case game.StateWaitOpponentReady:
			line2 = ui.GrayText.Render("Waiting for opponent to be ready")
		}

		if m.matchInfo.Mode == protocol.MULTI_REMOTE && m.postRoundTimer.Running() {
			seconds := int(m.postRoundTimer.Timeout.Seconds())
			countdown := fmt.Sprintf("%2d", seconds)

			line3 = lipgloss.JoinHorizontal(lipgloss.Center,
				ui.GrayText.Render("Next round will begin in "),
				ui.RoseText.Render(countdown),
				ui.GrayText.Render(" seconds"),
			)
		}

		switch m.matchInfo.Format {
		case protocol.WORDLE:
			if points == 0 {
				outcome = fmt.Sprintf(
					"%s Not guessed",
					ui.DrawBlock(),
				)

				line1 = lipgloss.JoinHorizontal(lipgloss.Center,
					outcome,
					" - ",
					strings.Join(words, " • "),
				)
			} else {
				line1 = outcome
			}
		case protocol.QUORDLE:
			if points == 0 {
				outcome = fmt.Sprintf(
					"%s Round draw",
					ui.DrawBlock(),
				)
			}

			line1 = lipgloss.JoinHorizontal(lipgloss.Center,
				outcome,
				" - ",
				strings.Join(words, " • "),
			)
		}
	} else {
		line1 = ui.SmallLogo()
		if m.err != nil {
			line2 = ui.RoseText.Render(fmt.Sprintf("Error: %s", m.err))
		}
	}

	content = lipgloss.JoinVertical(
		lipgloss.Center,
		line1,
		line2,
		line3,
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(2).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Top).
		Render(content)
}

func (m *gameModel) validateGuess(guess string) (bool, error) {
	repeatedGuessErr := fmt.Errorf("'%s' was already guessed", guess)
	invalidGuessErr := fmt.Errorf("'%s' is not a valid guess", guess)

	if m.alreadyGuessed(guess) {
		return false, repeatedGuessErr
	}

	switch len(guess) {
	case 5:
		return slices.Index(game.FiveLetterWords, guess) != -1, invalidGuessErr
	case 6:
		return slices.Index(game.SixLetterWords, guess) != -1, invalidGuessErr
	case 7:
		return slices.Index(game.SevenLetterWords, guess) != -1, invalidGuessErr
	default:
		logger.Error("[validateGuess] received value of length %d", len(guess))
		return false, invalidGuessErr
	}
}

func emit(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

func (m *gameModel) alreadyGuessed(guess string) bool {
	return slices.Contains(m.guesses, guess)
}

func (m *gameModel) addGuess(word string) {
	m.guesses[m.matchInfo.CurrentAttempt] = word
}

func (m *gameModel) isLastRound() bool {
	return m.matchInfo.CurrentRound >= m.matchInfo.TotalRounds
}

func (m *gameModel) initChallenges() {
	logger.Debug("initChallenges()")
	var challengesLen int
	switch m.matchInfo.Format {
	case protocol.WORDLE:
		challengesLen = 1
	case protocol.QUORDLE:
		challengesLen = 4
	}

	m.challengesLen = challengesLen
	m.guesses = make([]string, m.matchInfo.MaxAttempts)
	m.challenges = make([]*protocol.WordChallenge, challengesLen)

	for i := range challengesLen {
		m.challenges[i] = protocol.NewWordChallenge(m.matchInfo.MaxAttempts)
	}
}

func (m *gameModel) setTurnTimer() tea.Cmd {
	if m.turnTimeout > 0 {
		m.turnTimer = timer.New(time.Second * time.Duration(m.turnTimeout))
		return m.turnTimer.Init()
	}
	return nil
}

func (m *gameModel) setPostRoundTimer() tea.Cmd {
	if m.postRoundTimeout > 0 && !m.isLastRound() {
		m.postRoundTimer = timer.New(time.Second * time.Duration(m.postRoundTimeout))
		return m.postRoundTimer.Init()
	}
	return nil
}
