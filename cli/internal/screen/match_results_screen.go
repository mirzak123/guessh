package screen

import (
	"fmt"
	"guessh/internal/game"
	"guessh/internal/protocol"

	tea "github.com/charmbracelet/bubbletea"
)

type matchResultsModel struct {
	mode         protocol.GameMode
	roundsPlayed int
	roundsWon    int
	outcome      protocol.Outcome
}

func NewMatchResults(mode protocol.GameMode, roundsPlayed, roundsWon int, outcome protocol.Outcome) *matchResultsModel {
	return &matchResultsModel{
		mode:         mode,
		roundsPlayed: roundsPlayed,
		roundsWon:    roundsWon,
		outcome:      outcome,
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
		"Rounds played: %d\nRounds guessed correctly: %d",
		m.roundsPlayed,
		m.roundsWon,
	)

	if m.mode == protocol.MULTI_REMOTE {
		switch m.outcome {
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
