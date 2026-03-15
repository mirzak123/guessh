package game

import (
	"guessh/internal/protocol"
)

type GameState int

const (
	StateInit GameState = iota
	StateMatchStarted
	StateWaitGuess
	StateWaitingGuessResult
	StateWaitOpponentJoin
	StateWaitOpponentGuess
	StateRoundFinished
	StateMatchFinished
)

type MatchInfo struct {
	Mode                protocol.GameMode
	Format              protocol.GameFormat
	WordLen             int
	CurrentRound        int
	RawTotalRounds      string
	TotalRounds         int
	RoundsPlayed        int
	MaxAttempts         int
	CurrentAttempt      int
	RoundPoints         []int
	PlayerName          string
	OpponentName        string
	PlayerOnTurn        bool
	RoomID              string
	JoinExisting        bool
	RoomValidationError error
	DeniedRematch       bool
}

func NewMatchInfo() *MatchInfo {
	return &MatchInfo{}
}

type RoundInfo struct {
	Word   string
	Points int
}

func NewRoundInfo() *RoundInfo {
	return &RoundInfo{}
}

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
	case StateWaitingGuessResult:
		str = "StateWaitGuess"
	}
	return str
}
