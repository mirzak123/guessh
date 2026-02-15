package game

import "guessh/internal/protocol"

type PauseIntent struct{}

type ContinueIntent struct{}

type CreateMatchIntent struct {
	Mode    protocol.GameMode
	WordLen int
	Rounds  int
}

type MakeGuessIntent struct {
	Guess string
}

type StartGameIntent struct{}

type JoinRoom struct {
	RoomId string
}
