package protocol

import (
	"encoding/json"
	"log"
)

type EnvelopeMessage struct {
	Type string `json:"type"`
}

/* Client Types */

type CreateMatchMessage struct {
	Type       string   `json:"type"`
	Mode       GameMode `json:"mode"`
	Rounds     int      `json:"rounds"`
	WordLength int      `json:"wordLength"`
}

type MakeGuessMessage struct {
	Type  string `json:"type"`
	Guess string `json:"guess"`
}

func NewMakeGuessMessage(guess string) *MakeGuessMessage {
	return &MakeGuessMessage{
		Type:  "MAKE_GUESS",
		Guess: guess,
	}
}

func (m *MakeGuessMessage) JSON() string {
	if v, err := json.Marshal(m); err != nil {
		log.Print("[MakeGuessMessage.JSON] Failed to marshal")
		return ""
	} else {
		return string(v)
	}
}

type LeaveMatchMessage struct {
	EnvelopeMessage
}

/* Server Types */

type ErrorMessage struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type MatchStartedMessage struct {
	Type       string `json:"type"`
	MatchID    string `json:"matchId"`
	Rounds     int    `json:"rounds"`
	WordLength int    `json:"wordLength"`
}

type RoundStartedMessage struct {
	Type        string `json:"type"`
	RoundNumber int    `json:"roundNumber"`
}

type WaitGuessMessage struct {
	Type string `json:"type"`
}

type WaitOpponentGuessMessage struct {
	Type string `json:"type"`
}

type GuessResultMessage struct {
	Type     string           `json:"type"`
	Success  bool             `json:"success"`
	Feedback []LetterFeedback `json:"feedback"`
}

type RoundFinishedMessage struct {
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Word    string `json:"word"`
}

type MatchFinishedMessage struct {
	Type   string `json:"type"`
	Winner string `json:"winner"`
}
