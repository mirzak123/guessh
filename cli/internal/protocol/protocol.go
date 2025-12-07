package protocol

import "github.com/charmbracelet/lipgloss"

type GameMode string

const (
	SINGLE       GameMode = "SINGLE"
	MULTI_LOCAL  GameMode = "MULTI_LOCAL"
	MULTI_REMOTE GameMode = "MULTI_REMOTE"
)

type LetterFeedback int

const (
	LETTER_ABSENT LetterFeedback = iota
	LETTER_PRESENT
	LETTER_CORRECT
)

var (
	green  = "46"
	yellow = "226"
	gray   = "7"

	correctStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(green)).
			Foreground(lipgloss.Color(green)).
			Bold(true).
			Padding(0, 1)

	presentStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(yellow)).
			Foreground(lipgloss.Color(yellow)).
			Bold(true).
			Padding(0, 1)

	absentStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(gray)).
			Foreground(lipgloss.Color(gray)).
			Bold(true).
			Padding(0, 1)
)

type MatchInfo struct {
	Mode      GameMode
	WordLen   int
	RawRounds string
}

type Guess struct {
	Word   string
	Result []LetterFeedback
}

func NewGuess(word string, result []LetterFeedback) *Guess {
	return &Guess{
		Word:   word,
		Result: result,
	}
}

func (g *Guess) View() string {
	blocks := make([]string, len(g.Word))

	for i, r := range g.Word {
		var style lipgloss.Style

		switch g.Result[i] {
		case LETTER_CORRECT:
			style = correctStyle
		case LETTER_PRESENT:
			style = presentStyle
		case LETTER_ABSENT:
			style = absentStyle
		}

		blocks[i] = style.Render(string(r))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, blocks...)
}
