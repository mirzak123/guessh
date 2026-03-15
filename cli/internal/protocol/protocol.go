package protocol

type GameMode string
type GameFormat string
type LetterFeedback int
type Outcome int

const ROOM_ID_LENGTH = 5

const (
	SINGLE       GameMode = "SINGLE"
	MULTI_LOCAL  GameMode = "MULTI_LOCAL"
	MULTI_REMOTE GameMode = "MULTI_REMOTE"
)

const (
	WORDLE  GameFormat = "WORDLE"
	QUORDLE GameFormat = "QUORDLE"
)

const (
	LETTER_ABSENT LetterFeedback = iota
	LETTER_PRESENT
	LETTER_CORRECT
)

const (
	OUTCOME_NONE Outcome = iota
	OUTCOME_PLAYER_WON
	OUTCOME_OPPONENT_WON
)

type WordChallenge struct {
	CorrectWord  string
	SolvedBy     Outcome
	SolvedOnTurn int
	Feedbacks    [][]LetterFeedback
}

func NewWordChallenge(maxAttempts int) *WordChallenge {
	feedbacks := make([][]LetterFeedback, maxAttempts)

	return &WordChallenge{
		SolvedBy:  OUTCOME_NONE,
		Feedbacks: feedbacks,
	}
}
