package screen

import (
	"fmt"
	"guessh/internal/game"
	"guessh/internal/protocol"
	"guessh/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type matchResultsModel struct {
	mode         protocol.GameMode
	format       protocol.GameFormat
	roundsPlayed int
	roundPoints  []int
	matchOutcome protocol.Outcome
	canRematch   bool
	confirm      bool

	form *huh.Form
}

func NewMatchResults(
	mode protocol.GameMode,
	format protocol.GameFormat,
	roundsPlayed int,
	roundPoints []int,
	matchOutcome protocol.Outcome,
	playerName string,
	opponentName string,
	opponentLeft bool) *matchResultsModel {

	m := &matchResultsModel{
		mode:         mode,
		format:       format,
		roundsPlayed: roundsPlayed,
		roundPoints:  roundPoints,
		matchOutcome: matchOutcome,
		canRematch:   !opponentLeft,
		confirm:      true,
	}

	var confirmInput *huh.Confirm
	var summary string

	results := lipgloss.JoinHorizontal(
		lipgloss.Left,
		"Round outcomes • ",
		ui.ViewRoundOutcomes(m.roundPoints, m.format, m.roundsPlayed))

	switch m.mode {
	case protocol.MULTI_REMOTE:
		if opponentLeft {
			summary = "🔌 Opponent left the match"
		} else {
			switch m.matchOutcome {
			case protocol.OUTCOME_PLAYER_WON:
				summary = "🎉 You won!"
			case protocol.OUTCOME_OPPONENT_WON:
				summary = "😢 You lost"
			default:
				summary = "🤝 Draw"
			}
		}
	case protocol.MULTI_LOCAL:
		switch m.matchOutcome {
		case protocol.OUTCOME_PLAYER_WON:
			summary = fmt.Sprintf("🏅 %s won!", ui.PurpleText.Render(playerName))
		case protocol.OUTCOME_OPPONENT_WON:
			summary = fmt.Sprintf("🏅 %s won!", ui.RoseText.Render(opponentName))
		case protocol.OUTCOME_NONE:
			summary = "🤝 Draw"
		}
	}

	summary = lipgloss.JoinVertical(
		lipgloss.Left,
		summary,
		results,
	)

	if m.canRematch {
		confirmInput = huh.NewConfirm().
			Title("Rematch?").
			Affirmative("Request rematch").
			Negative("Continue to main screen").
			Value(&m.confirm).
			WithButtonAlignment(lipgloss.Left)
	} else {
		confirmInput = huh.NewConfirm().
			Title("Continue to Main Screen").
			Affirmative("Continue").
			Negative("").
			Value(&m.confirm).
			WithButtonAlignment(lipgloss.Left)
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Match Results").
				Description(summary),
			confirmInput,
		),
	).WithShowHelp(false)

	m.form.NextField()

	return m
}

func (m matchResultsModel) Init() tea.Cmd {
	return nil
}

func (m matchResultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc {
			return m, nil
		}
	}

	updatedModel, cmd := m.form.Update(msg)
	m.form = updatedModel.(*huh.Form)

	if m.form.State == huh.StateCompleted {
		if m.confirm {
			if m.canRematch {
				return m, emit(game.RequestRematchIntent{})
			} else {
				return m, tea.Batch(
					emit(game.StartMenuIntent{}),
				)
			}
		} else {
			return m, emit(game.DenyRematchIntent{})
		}
	}

	return m, cmd
}

func (m matchResultsModel) View() string {
	return m.form.View()
}
