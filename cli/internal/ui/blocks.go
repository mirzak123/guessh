package ui

import (
	"guessh/internal/protocol"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

func ViewWordInputRow(input string, length int, isActive bool) string {
	blocks := make([]string, length)
	var style lipgloss.Style

	if isActive {
		style = activeInputStyle
	} else {
		style = baseLetterStyle
	}

	for i := range blocks {
		var char string
		if i < len(input) {
			char = string(input[i])
		} else {
			char = " "
		}

		blocks[i] = style.Render(char)
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, blocks...)
}

func ViewGuessGrid(guesses []*protocol.Guess, input string, maxAttempts int, wordLen int, onTurn bool) string {
	grid := make([]string, maxAttempts)

	for i := range maxAttempts {
		if i < len(guesses) {
			grid[i] = ViewGuessedRow(guesses[i])
		} else if i == len(guesses) {
			grid[i] = ViewWordInputRow(input, wordLen, onTurn)
		} else {
			grid[i] = ViewWordInputRow("", wordLen, false)
		}
	}

	return strings.Join(grid, "\n")
}
