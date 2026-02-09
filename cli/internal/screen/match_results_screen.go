package screen

import (
	"fmt"
	"guessh/internal/logger"

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
	logger.Debug("Update(%v)", msg)
	switch msg.(type) {
	case tea.KeyMsg:
		return m, func() tea.Msg { return StartGameMsg{} }
	}
	return m, nil
}

func (m matchResultsModel) View() string {
	logger.Debug("View()")
	return fmt.Sprintf(
		"Rounds played: %d\nRounds guessed correctly: %d",
		m.roundsPlayed,
		m.roundsWon,
	)
}
