package screen

import (
	"guessh/internal/game"
	"guessh/internal/protocol"
	"guessh/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type matchResultsModel struct {
	mode          protocol.GameMode
	roundsPlayed  int
	roundOutcomes []*protocol.Outcome
	matchOutcome  protocol.Outcome
	confirm       bool

	form *huh.Form
}

func NewMatchResults(
	mode protocol.GameMode,
	roundsPlayed int,
	roundOutcomes []*protocol.Outcome,
	matchOutcome protocol.Outcome,
	opponentLeft bool) *matchResultsModel {

	m := &matchResultsModel{
		mode:          mode,
		roundsPlayed:  roundsPlayed,
		roundOutcomes: roundOutcomes,
		matchOutcome:  matchOutcome,
		confirm:       true,
	}

	var confirmInput *huh.Confirm
	var results string

	switch m.mode {
	case protocol.MULTI_REMOTE:
		if opponentLeft {
			results = "🔌 Opponent left the match"
		} else {
			switch m.matchOutcome {
			case protocol.OUTCOME_PLAYER_WON:
				results = "🎉 You won!"
			case protocol.OUTCOME_OPPONENT_WON:
				results = "😢 You lost"
			default:
				results = "🤝 Draw"
			}
		}

		results = lipgloss.JoinVertical(
			lipgloss.Left,
			results,
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				"Round outcomes • ",
				ui.ViewRoundOutcomes(m.roundOutcomes)),
		)

		confirmInput = huh.NewConfirm().
			Title("Rematch?").
			Affirmative("Request rematch").
			Negative("Continue to main screen").
			Value(&m.confirm).
			WithButtonAlignment(lipgloss.Left)
	case protocol.SINGLE:
		results = lipgloss.JoinHorizontal(
			lipgloss.Left,
			"Round outcomes • ",
			ui.ViewRoundOutcomes(m.roundOutcomes))

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
				Description(results),
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
			switch m.mode {

			case protocol.MULTI_REMOTE:
				return m, emit(game.RequestRematchIntent{})

			case protocol.SINGLE:
				return m, tea.Batch(
					emit(game.StartGameIntent{}),
				)

			}
		} else {
			// only MULTI_REMOTE can get here
			return m, tea.Batch(
				emit(game.DenyRematchIntent{}),
				emit(game.StartGameIntent{}),
			)
		}
	}

	return m, cmd
}

func (m matchResultsModel) View() string {
	return m.form.View()
}
