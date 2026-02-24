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

	form *huh.Form
}

func NewMatchResults(mode protocol.GameMode, roundsPlayed int, roundOutcomes []*protocol.Outcome, matchOutcome protocol.Outcome) *matchResultsModel {
	m := &matchResultsModel{
		mode:          mode,
		roundsPlayed:  roundsPlayed,
		roundOutcomes: roundOutcomes,
		matchOutcome:  matchOutcome,
	}

	var results string
	if m.mode == protocol.MULTI_REMOTE {
		switch m.matchOutcome {
		case protocol.OUTCOME_PLAYER_WON:
			results = "🎉 You won!"
		case protocol.OUTCOME_OPPONENT_WON:
			results = "😢 You lost"
		default:
			results = "🤝 Draw"
		}

		results = lipgloss.JoinVertical(
			lipgloss.Top,
			results,
			ui.ViewRoundOutcomes(m.roundOutcomes))
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Match Results").
				Description(results),
			huh.NewConfirm().
				Title("Continue to Main Screen").
				Affirmative("Continue").
				Negative("").
				WithButtonAlignment(lipgloss.Left),
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
		return m, emit(game.StartGameIntent{})
	}

	return m, cmd
}

func (m matchResultsModel) View() string {
	return m.form.View()
}
