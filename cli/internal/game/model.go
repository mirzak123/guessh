package game

import (
	"guessh/internal/protocol"
	"slices"
)

type GameState int

const (
	StateInit GameState = iota
	StateMatchStarted
	StateWaitGuess
	StateWaitOpponentReady
	StateWaitOpponentJoin
	StateWaitOpponentGuess
	StateRoundFinished
	StateMatchFinished
)

func (s GameState) String() string {
	var str string
	switch s {
	case StateInit:
		str = "StateInit"
	case StateMatchFinished:
		str = "StateMatchFinished"
	case StateMatchStarted:
		str = "StateMatchStarted"
	case StateRoundFinished:
		str = "StateRoundFinished"
	case StateWaitGuess:
		str = "StateWaitGuess"
	case StateWaitOpponentGuess:
		str = "StateWaitOpponentGuess"
	case StateWaitOpponentJoin:
		str = "StateWaitOpponentJoin"
	case StateWaitOpponentReady:
		str = "StateWaitOpponentReady"
	}
	return str
}

type MatchInfo struct {
	Mode                protocol.GameMode
	Format              protocol.GameFormat
	WordLen             int
	CurrentRound        int
	RawTotalRounds      string
	TotalRounds         int
	TurnTimeout         int
	RoundsPlayed        int
	MaxAttempts         int
	CurrentAttempt      int
	CorrectWords        []string
	RoundPoints         []int
	PlayerName          string
	OpponentName        string
	PlayerOnTurn        bool
	RoomID              string
	JoinExisting        bool
	RoomValidationError error
	DeniedRematch       bool
	Guesses             []string
	Challenges          []*protocol.WordChallenge
}

func NewMatchInfo() *MatchInfo {
	return &MatchInfo{}
}

func (m *MatchInfo) AlreadyGuessed(guess string) bool {
	return slices.Contains(m.Guesses, guess)
}

func (m *MatchInfo) AddGuess(word string) {
	m.Guesses[m.CurrentAttempt] = word
}

func (m *MatchInfo) IsLastRound() bool {
	return m.CurrentRound >= m.TotalRounds
}
