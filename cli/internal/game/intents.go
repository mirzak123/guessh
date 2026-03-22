package game

import (
	"guessh/internal/protocol"
)

type ContinueIntent struct{}

type CreateMatchIntent struct {
	Mode           protocol.GameMode
	Format         protocol.GameFormat
	WordLen        int
	Rounds         int
	SecondsPerTurn int
	PlayerName     string
}

type MakeGuessIntent struct {
	Guess string
}

type StartMenuIntent struct{}

type PlayGameIntent struct{}

type TypingIntent struct {
	Value string
}

type JoinRoomIntent struct {
	RoomId     string
	PlayerName string
}

type LeaveMatchIntent struct{}

type RequestRematchIntent struct{}

type DenyRematchIntent struct{}
