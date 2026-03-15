package ui

import (
	"fmt"
	"guessh/internal/protocol"

	"github.com/charmbracelet/lipgloss"
)

func PlayerBlock() string {
	return viewSymbolBlock(symbolCheck, White, Purple)
}

func OpponentBlock() string {
	return viewSymbolBlock(symbolCross, White, Rose)
}

func DrawBlock() string {
	return viewSymbolBlock(symbolDraw, White, Gray)
}

func NoOpponentBlock() string {
	return viewSymbolBlock(symbolNone, White, "")
}

func RoundNotPlayedBlock() string {
	return viewSymbolBlock(symbolDraw, White, "")
}

func PointOutcomeBlock(points int) string {
	var bg lipgloss.Color
	if points > 0 {
		bg = Purple
	} else if points < 0 {
		points *= -1
		bg = Rose
	} else {
		bg = Gray
	}
	return viewSymbolBlock(fmt.Sprintf("%d", points), White, bg)
}

func RoundOutcomeBlock(points int, format protocol.GameFormat) string {
	if format == protocol.WORDLE {
		if points > 0 {
			return PlayerBlock()
		}
		if points < 0 {
			return OpponentBlock()
		}
		return DrawBlock()
	}

	return PointOutcomeBlock(points)
}

func viewSymbolBlock(symbol string, fg, bg lipgloss.Color) string {
	style := lipgloss.NewStyle().
		Width(3).
		Height(1).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Foreground(fg).
		Bold(true)

	if bg != "" {
		style = style.Background(bg)
	}

	return style.Render(symbol)
}
