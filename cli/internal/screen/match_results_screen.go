package screen

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type matchResultsModel struct {
	roundsPlayed int
	roundsWon    int
}

func NewMatchResults(roundsPlayed, roundsWon int) matchResultsModel {
	return matchResultsModel{
		roundsPlayed: roundsPlayed,
		roundsWon:    roundsWon,
	}
}

func (m matchResultsModel) Init() tea.Cmd {
	return nil
}

func (m matchResultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, func() tea.Msg { return StartGameMsg{} }
	}
	return m, nil
}

func (m matchResultsModel) View() string {
	return fmt.Sprintf(
		"Rounds played: %d\nRounds guessed correctly: %d",
		m.roundsPlayed,
		m.roundsWon,
	)
}
