package ui

import (
	"guessh/internal/game"
	"guessh/internal/protocol"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func ViewGuessedRow(g *protocol.Guess) string {
	blocks := make([]string, len(g.Result))

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

func ViewWordInputRow(input string, length int, onTurn bool) string {
	blocks := make([]string, length)
	var style lipgloss.Style

	if onTurn {
		style = playerActiveInputStyle
	} else {
		style = opponentActiveInputStyle
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

func ViewInactiveRow(wordLen int) string {
	blocks := make([]string, wordLen)
	for i := range blocks {
		blocks[i] = baseLetterStyle.Render(" ")
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, blocks...)
}

func ViewGuessGrid(guesses []*protocol.Guess, input string, maxAttempts int, wordLen int, state game.GameState) string {
	grid := make([]string, maxAttempts)

	for i := range maxAttempts {
		if i < len(guesses) {
			grid[i] = ViewGuessedRow(guesses[i])
		} else if i == len(guesses) && (state == game.StateWaitGuess || state == game.StateWaitOpponentGuess) {
			grid[i] = ViewWordInputRow(input, wordLen, state == game.StateWaitGuess)
		} else {
			grid[i] = ViewInactiveRow(wordLen)
		}
	}

	return strings.Join(grid, "\n")
}

func ViewOutcomeBlock(points int, done, isLast bool) string {
	outcomeBlock := OutcomeBlock(points, done)

	if !isLast {
		outcomeBlock = lipgloss.NewStyle().MarginRight(1).Render(outcomeBlock)
	}

	return outcomeBlock
}

func ViewRoundOutcomes(pointsPerRound []int, roundsPlayed int) string {
	blocks := make([]string, len(pointsPerRound))
	for i, points := range pointsPerRound {
		blocks[i] = ViewOutcomeBlock(points, i < roundsPlayed, len(pointsPerRound)-1 == i)
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, blocks...)
}
