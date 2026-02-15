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
	Mode           protocol.GameMode
	WordLen        int
	CurrentRound   int
	RawTotalRounds string
	TotalRounds    int
	MaxAttempts    int
	RoundsWon      int
	PlayerName     string
	RoomID         string
	JoinExisting   bool
}

func NewMatchInfo() *MatchInfo {
	return &MatchInfo{
		CurrentRound: 0,
	}
}

type RoundInfo struct {
	Word    string
	Success bool
}

func NewRoundInfo() *RoundInfo {
	return &RoundInfo{}
}
