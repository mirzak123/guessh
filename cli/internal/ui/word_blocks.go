package ui

import (
	"guessh/internal/game"
	"guessh/internal/protocol"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func ViewGuessedRow(word string, result []protocol.LetterFeedback) string {
	blocks := make([]string, len(result))

	for i, r := range word {
		var style lipgloss.Style

		switch result[i] {
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

func ViewGuessGrid(guesses []string, challenge *protocol.WordChallenge, input string, currentAttempt, maxAttempts, wordLen int, state game.GameState) string {
	grid := make([]string, maxAttempts)

	for i := range maxAttempts {
		if i < currentAttempt {
			guess := guesses[i]
			grid[i] = ViewGuessedRow(guess, challenge.Feedbacks[i])
		} else if i == currentAttempt && (state == game.StateWaitGuess || state == game.StateWaitOpponentGuess) {
			grid[i] = ViewWordInputRow(input, wordLen, state == game.StateWaitGuess)
		} else {
			grid[i] = ViewInactiveRow(wordLen)
		}
	}

	return strings.Join(grid, "\n")
}

func ViewOutcomeBlock(points int, format protocol.GameFormat, done, isLast bool) string {
	var outcomeBlock string

	if done {
		outcomeBlock = RoundOutcomeBlock(points, format)
	} else {
		outcomeBlock = RoundNotPlayedBlock()
	}

	if !isLast {
		outcomeBlock = lipgloss.NewStyle().MarginRight(1).Render(outcomeBlock)
	}

	return outcomeBlock
}

func ViewRoundOutcomes(pointsPerRound []int, format protocol.GameFormat, roundsPlayed int) string {
	blocks := make([]string, len(pointsPerRound))
	for i, points := range pointsPerRound {
		blocks[i] = ViewOutcomeBlock(points, format, i < roundsPlayed, len(pointsPerRound)-1 == i)
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, blocks...)
}
