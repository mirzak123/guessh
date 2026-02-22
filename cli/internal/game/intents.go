package game

import "guessh/internal/protocol"

type PauseIntent struct{}

type ContinueIntent struct{}

type CreateMatchIntent struct {
	Mode       protocol.GameMode
	WordLen    int
	Rounds     int
	PlayerName string
}

type MakeGuessIntent struct {
	Guess string
}

type StartGameIntent struct{}

type TypingIntent struct {
	Value string
}

type JoinRoomIntent struct {
	RoomId     string
	PlayerName string
}
