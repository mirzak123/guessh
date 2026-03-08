package ui

import (
	"guessh/internal/protocol"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	Green  = lipgloss.Color("46")
	Yellow = lipgloss.Color("226")
	Gray   = lipgloss.Color("244")
	White  = lipgloss.Color("15")
	Purple = lipgloss.Color("63")
	Rose   = lipgloss.Color("198")
)

var (
	Theme = huh.ThemeCharm()

	MainContentStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Purple).
				Padding(1, 2)

	baseLetterStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Bold(true).
			Padding(0, 1)

	correctStyle = baseLetterStyle.
			BorderForeground(Green).
			Foreground(Green)

	presentStyle = baseLetterStyle.
			BorderForeground(Yellow).
			Foreground(Yellow)

	absentStyle = baseLetterStyle.
			BorderForeground(Gray).
			Foreground(Gray)

	playerActiveInputStyle = baseLetterStyle.
				BorderForeground(Purple).
				Foreground(White)

	opponentActiveInputStyle = baseLetterStyle.
					BorderForeground(Rose).
					Foreground(White)

	PurpleText = lipgloss.NewStyle().
			Foreground(Purple)

	RoseText = lipgloss.NewStyle().
			Foreground(Rose)

	GrayText = lipgloss.NewStyle().
			Foreground(Gray)

	WhiteText = lipgloss.NewStyle().
			Foreground(White)
)

func OutcomeBlock(outcome *protocol.Outcome) string {
	var (
		fg     = White
		bg     lipgloss.Color
		symbol string
		withBg = true
	)

	if outcome == nil {
		withBg = false
		symbol = "○"
	} else {
		switch *outcome {
		case protocol.OUTCOME_PLAYER_WON:
			bg = Purple
			symbol = "✓"
		case protocol.OUTCOME_OPPONENT_WON:
			bg = Rose
			symbol = "✗"
		case protocol.OUTCOME_NONE:
			bg = Gray
			symbol = "○"
		}
	}

	style := lipgloss.NewStyle().
		Width(3).
		Height(1).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Foreground(fg).
		Bold(true)

	if withBg {
		style = style.Background(bg)
	}

	return style.Render(symbol)
}
