package protocol

type EventType string

const (
	CREATE_MATCH            EventType = "CREATE_MATCH"
	JOIN_ROOM               EventType = "JOIN_ROOM"
	MAKE_GUESS              EventType = "MAKE_GUESS"
	LEAVE_MATCH             EventType = "LEAVE_MATCH"
	ERROR                   EventType = "ERROR"
	ROOM_CREATED            EventType = "ROOM_CREATED"
	ROOM_JOINED             EventType = "ROOM_JOINED"
	ROOM_JOIN_FAILED        EventType = "ROOM_JOIN_FAILED"
	WAIT_OPPONENT_JOIN      EventType = "WAIT_OPPONENT_JOIN"
	REQUEST_REMATCH         EventType = "REQUEST_REMATCH"
	DENY_REMATCH            EventType = "DENY_REMATCH"
	OPPONENT_DENIED_REMATCH EventType = "OPPONENT_DENIED_REMATCH"
	OPPONENT_LEFT           EventType = "OPPONENT_LEFT"
	MATCH_STARTED           EventType = "MATCH_STARTED"
	ROUND_STARTED           EventType = "ROUND_STARTED"
	WAIT_GUESS              EventType = "WAIT_GUESS"
	WAIT_OPPONENT_GUESS     EventType = "WAIT_OPPONENT_GUESS"
	GUESS_RESULT            EventType = "GUESS_RESULT"
	ROUND_FINISHED          EventType = "ROUND_FINISHED"
	MATCH_FINISHED          EventType = "MATCH_FINISHED"
	TYPING                  EventType = "TYPING"
	OPPONENT_TYPING         EventType = "OPPONENT_TYPING"
)

type EnvelopeEvent struct {
	Type EventType `json:"type"`
}

/* Client Types */

type CreateMatchEvent struct {
	Type       EventType  `json:"type"`
	Mode       GameMode   `json:"mode"`
	Format     GameFormat `json:"format"`
	WordLen    int        `json:"wordLength"`
	Rounds     int        `json:"rounds"`
	PlayerName string     `json:"playerName,omitempty"`
}

func NewCreateMatchEvent(mode GameMode, format GameFormat, wordLen int, rounds int, playerName string) *CreateMatchEvent {
	return &CreateMatchEvent{
		Type:       CREATE_MATCH,
		Mode:       mode,
		Format:     format,
		WordLen:    wordLen,
		Rounds:     rounds,
		PlayerName: playerName,
	}
}

type MakeGuessEvent struct {
	Type  EventType `json:"type"`
	Guess string    `json:"guess"`
}

func NewMakeGuessEvent(guess string) *MakeGuessEvent {
	return &MakeGuessEvent{
		Type:  MAKE_GUESS,
		Guess: guess,
	}
}

type LeaveMatchEvent struct {
	Type EventType `json:"type"`
}

func NewLeaveMatchEvent() *LeaveMatchEvent {
	return &LeaveMatchEvent{
		Type: LEAVE_MATCH,
	}
}

type JoinRoomEvent struct {
	Type       EventType `json:"type"`
	RoomID     string    `json:"roomId"`
	PlayerName string    `json:"playerName,omitempty"`
}

func NewJoinRoomEvent(roomID string, playerName string) *JoinRoomEvent {
	return &JoinRoomEvent{
		Type:       JOIN_ROOM,
		RoomID:     roomID,
		PlayerName: playerName,
	}
}

type RequestRematchEvent struct {
	Type EventType `json:"type"`
}

func NewRequestRematchEvent() *RequestRematchEvent {
	return &RequestRematchEvent{
		Type: REQUEST_REMATCH,
	}
}

type DenyRematchEvent struct {
	Type EventType `json:"type"`
}

func NewDenyRematchEvent() *DenyRematchEvent {
	return &DenyRematchEvent{
		Type: DENY_REMATCH,
	}
}

type TypingEvent struct {
	Type  EventType `json:"type"`
	Value string    `json:"value"`
}

func NewTypingEvent(value string) *TypingEvent {
	return &TypingEvent{
		Type:  TYPING,
		Value: value,
	}
}

/* Server Types */

type ErrorEvent struct {
	Type   EventType `json:"type"`
	Reason string    `json:"reason"`
}

type MatchStartedEvent struct {
	Type         EventType  `json:"type"`
	MatchID      string     `json:"matchId"`
	Format       GameFormat `json:"format"`
	Rounds       int        `json:"rounds"`
	WordLength   int        `json:"wordLength"`
	OpponentName string     `json:"opponentName,omitempty"`
}

type RoundStartedEvent struct {
	Type        EventType `json:"type"`
	RoundNumber int       `json:"roundNumber"`
	MaxAttempts int       `json:"maxAttempts"`
}

type WaitGuessEvent struct {
	Type EventType `json:"type"`
}

type WaitOpponentGuessEvent struct {
	Type EventType `json:"type"`
}

type GuessResultEvent struct {
	Type     EventType          `json:"type"`
	Guess    string             `json:"guess"`
	Feedback [][]LetterFeedback `json:"feedback"`
}

type RoundFinishedEvent struct {
	Type   EventType `json:"type"`
	Points int       `json:"points"`
	Words  []string  `json:"words"`
}

type MatchFinishedEvent struct {
	Type         EventType `json:"type"`
	Outcome      Outcome   `json:"outcome"`
	OpponentLeft bool      `json:"opponentLeft"`
}

type RoomCreatedEvent struct {
	Type   EventType `json:"type"`
	RoomID string    `json:"roomId"`
}

type RoomJoinFailedEvent struct {
	Type   EventType `json:"type"`
	RoomID string    `json:"roomId"`
	Reason string    `json:"reason"`
}

type OpponentTypingEvent struct {
	Type  EventType `json:"type"`
	Value string    `json:"value"`
}
