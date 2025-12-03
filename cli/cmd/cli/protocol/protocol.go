package protocol

type GameMode string

const (
	SINGLE       = "SINGLE"
	MULTI_LOCAL  = "MULTI_LOCAL"
	MULTI_REMOTE = "MULTI_REMOTE"
)

type LetterFeedback int

const (
	LETTER_ABSENT  = 0
	LETTER_PRESENT = 1
	LETTER_CORRECT = 2
)
