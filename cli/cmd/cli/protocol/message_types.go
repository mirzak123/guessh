package protocol

type Envelope struct {
	Type string `json:"type"`
}

/* Client Types */

type CreateMatch struct {
	Envelope
	Mode       GameMode `json:"mode"`
	Rounds     int      `json:"rounds"`
	WordLength int      `json:"wordLength"`
}

type MakeGuess struct {
	Envelope
	Guess string `json:"guess"`
}

type LeaveMatch struct {
	Envelope
}

/* Server Types */

type Error struct {
	Envelope
	Reason string `json:"reason"`
}

type MatchStarted struct {
	Envelope
	MatchID    string `json:"matchId"`
	Rounds     int    `json:"rounds"`
	WordLength int    `json:"wordLength"`
}

type RoundStarted struct {
	Envelope
	RoundNumber int `json:"roundNumber"`
}

type WaitGuess struct {
	Envelope
}

type WaitOpponentGuess struct {
	Envelope
}

type GuessResult struct {
	Envelope
	Success  bool             `json:"success"`
	Feedback []LetterFeedback `json:"feedback"`
}

type RoundFinished struct {
	Envelope
	Success bool   `json:"success"`
	Word    string `json:"word"`
}

type MatchFinished struct {
	Envelope
	Winner string `json:"winner"`
}
