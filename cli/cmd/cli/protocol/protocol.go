package protocol

type GameMode string

const (
	SINGLE       GameMode = "SINGLE"
	MULTI_LOCAL  GameMode = "MULTI_LOCAL"
	MULTI_REMOTE GameMode = "MULTI_REMOTE"
)

type LetterFeedback int

const (
	LETTER_ABSENT  LetterFeedback = 0
	LETTER_PRESENT LetterFeedback = 1
	LETTER_CORRECT LetterFeedback = 2
)

type MatchInfo struct {
	Mode      GameMode
	WordLen   int
	RawRounds string
}
