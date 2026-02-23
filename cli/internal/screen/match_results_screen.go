package screen

import (
	"fmt"
	"guessh/internal/game"
	"guessh/internal/protocol"
	"guessh/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

type matchResultsModel struct {
	mode          protocol.GameMode
	roundsPlayed  int
	roundOutcomes []*protocol.Outcome
	matchOutcome  protocol.Outcome
}

func NewMatchResults(mode protocol.GameMode, roundsPlayed int, roundOutcomes []*protocol.Outcome, matchOutcome protocol.Outcome) *matchResultsModel {
	return &matchResultsModel{
		mode:          mode,
		roundsPlayed:  roundsPlayed,
		roundOutcomes: roundOutcomes,
		matchOutcome:  matchOutcome,
	}
}

func (m matchResultsModel) Init() tea.Cmd {
	return nil
}

func (m matchResultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, emit(game.StartGameIntent{})
	}
	return m, nil
}

func (m matchResultsModel) View() string {
	view := fmt.Sprintf(
		"Round %d/%d\n%s",
		m.roundsPlayed,
		len(m.roundOutcomes),
		ui.ViewRoundOutcomes(m.roundOutcomes),
	)

	if m.mode == protocol.MULTI_REMOTE {
		switch m.matchOutcome {
		case protocol.OUTCOME_PLAYER_WON:
			view += "\nYou won!"
		case protocol.OUTCOME_OPPONENT_WON:
			view += "\nYou lost :("
		default:
			view += "\nIt was a draw"
		}
	}
	return view
}
