package protocol

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
