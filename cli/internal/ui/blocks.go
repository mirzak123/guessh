package ui

import (
	"guessh/internal/protocol"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	green  = "46"
	yellow = "226"
	gray   = "7"

	baseLetterStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Bold(true).
			Padding(0, 1)

	correctStyle = baseLetterStyle.
			BorderForeground(lipgloss.Color(green)).
			Foreground(lipgloss.Color(green))

	presentStyle = baseLetterStyle.
			BorderForeground(lipgloss.Color(yellow)).
			Foreground(lipgloss.Color(yellow))

	absentStyle = baseLetterStyle.
			BorderForeground(lipgloss.Color(gray)).
			Foreground(lipgloss.Color(gray))
)

func ViewGuessedRow(g *protocol.Guess) string {
	blocks := make([]string, len(g.Word))

	for i, r := range g.Word {
		var style lipgloss.Style

		switch g.Result[i] {
		case protocol.LETTER_CORRECT:
			style = correctStyle
		case protocol.LETTER_PRESENT:
			style = presentStyle
		case protocol.LETTER_ABSENT:
			style = absentStyle
		}

		blocks[i] = style.Render(string(r))
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, blocks...)
}

func ViewWordInputRow(input string, length int) string {
	blocks := make([]string, length)

	for i := range blocks {
		var char string
		if i < len(input) {
			char = string(input[i])
		} else {
			char = " "
		}
		blocks[i] = baseLetterStyle.Render(char)
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, blocks...)
}

func ViewGuessGrid(guesses []*protocol.Guess, input string, maxAttempts int, wordLen int) string {
	grid := make([]string, maxAttempts)

	for i := range maxAttempts {
		if i < len(guesses) {
			grid[i] = ViewGuessedRow(guesses[i])
		} else if i == len(guesses) {
			grid[i] = ViewWordInputRow(input, wordLen)
		} else {
			grid[i] = ViewWordInputRow("", wordLen)
		}
	}

	return strings.Join(grid, "\n")
}
