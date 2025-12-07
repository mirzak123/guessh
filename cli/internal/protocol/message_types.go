package protocol

type MessageType string

const (
	CREATE_MATCH        MessageType = "CREATE_MATCH"
	JOIN_ROOM           MessageType = "JOIN_ROOM"
	CREATE_ROOM         MessageType = "CREATE_ROOM"
	MAKE_GUESS          MessageType = "MAKE_GUESS"
	REQUEST_REMATCH     MessageType = "REQUEST_REMATCH"
	LEAVE_MATCH         MessageType = "LEAVE_MATCH"
	ERROR               MessageType = "ERROR"
	CONNECTED           MessageType = "CONNECTED"
	ROOM_CREATED        MessageType = "ROOM_CREATED"
	ROOM_JOINED         MessageType = "ROOM_JOINED"
	ROOM_JOIN_FAILED    MessageType = "ROOM_JOIN_FAILED"
	WAIT_OPPONENT_JOIN  MessageType = "WAIT_OPPONENT_JOIN"
	MATCH_STARTED       MessageType = "MATCH_STARTED"
	ROUND_STARTED       MessageType = "ROUND_STARTED"
	WAIT_GUESS          MessageType = "WAIT_GUESS"
	WAIT_OPPONENT_GUESS MessageType = "WAIT_OPPONENT_GUESS"
	GUESS_RESULT        MessageType = "GUESS_RESULT"
	ROUND_FINISHED      MessageType = "ROUND_FINISHED"
	MATCH_FINISHED      MessageType = "MATCH_FINISHED"
	BYE                 MessageType = "BYE"
)

type EnvelopeMessage struct {
	Type MessageType `json:"type"`
}

/* Client Types */

type CreateMatchMessage struct {
	Type    MessageType `json:"type"`
	Mode    GameMode    `json:"mode"`
	WordLen int         `json:"wordLength"`
	Rounds  int         `json:"rounds"`
}

func NewCreateMatchMessage(mode GameMode, wordLen int, rounds int) *CreateMatchMessage {
	return &CreateMatchMessage{
		Type:    CREATE_MATCH,
		Mode:    mode,
		WordLen: wordLen,
		Rounds:  rounds,
	}
}

type MakeGuessMessage struct {
	Type  MessageType `json:"type"`
	Guess string      `json:"guess"`
}

func NewMakeGuessMessage(guess string) *MakeGuessMessage {
	return &MakeGuessMessage{
		Type:  MAKE_GUESS,
		Guess: guess,
	}
}

type LeaveMatchMessage struct {
	Type MessageType `json:"type"`
}

/* Server Types */

type ErrorMessage struct {
	Type   MessageType `json:"type"`
	Reason string      `json:"reason"`
}

type MatchStartedMessage struct {
	Type       MessageType `json:"type"`
	MatchID    string      `json:"matchId"`
	Rounds     int         `json:"rounds"`
	WordLength int         `json:"wordLength"`
}

type RoundStartedMessage struct {
	Type        MessageType `json:"type"`
	RoundNumber int         `json:"roundNumber"`
	MaxAttempts int         `json:"maxAttempts"`
}

type WaitGuessMessage struct {
	Type MessageType `json:"type"`
}

type WaitOpponentGuessMessage struct {
	Type MessageType `json:"type"`
}

type GuessResultMessage struct {
	Type     MessageType      `json:"type"`
	Success  bool             `json:"success"`
	Guess    string           `json:"guess"`
	Feedback []LetterFeedback `json:"feedback"`
}

type RoundFinishedMessage struct {
	Type    MessageType `json:"type"`
	Success bool        `json:"success"`
	Word    string      `json:"word"`
}

type MatchFinishedMessage struct {
	Type   MessageType `json:"type"`
	Winner string      `json:"winner"`
}
