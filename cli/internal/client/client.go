package client

import (
	"encoding/json"
	"guessh/internal/logger"
	"guessh/internal/protocol"
	"guessh/internal/transport"
	"net"
	"os"
)

type Client struct {
	Conn net.Conn
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		Conn: conn,
	}
}

func (c *Client) CreateMatch(mode protocol.GameMode, wordLen, rounds int, playerName string) {
	var (
		msg []byte
		err error
	)

	createMatchMsg := protocol.NewCreateMatchEvent(mode, wordLen, rounds, playerName)
	if msg, err = json.Marshal(createMatchMsg); err != nil {
		logger.Error("[Client.CreateMatch] Failed to marshal CreateMatchEvent: %v", err)
		os.Exit(1)
	}

	c.send(msg)
}

func (c *Client) MakeGuess(guess string) {
	var (
		msg []byte
		err error
	)

	makeGuessMsg := protocol.NewMakeGuessEvent(guess)
	if msg, err = json.Marshal(makeGuessMsg); err != nil {
		logger.Error("[Client.MakeGuess] Failed to marshal MakeGuessEvent: %v", err)
		os.Exit(1)
	}

	c.send(msg)
}

func (c *Client) JoinRoom(roomID string, playerName string) {
	var (
		msg []byte
		err error
	)

	joinRoomMsg := protocol.NewJoinRoomEvent(roomID, playerName)
	if msg, err = json.Marshal(joinRoomMsg); err != nil {
		logger.Error("[Client.MakeGuess] Failed to marshal JoinRoomEvent: %v", err)
		os.Exit(1)
	}

	c.send(msg)
}

func (c *Client) Typing(value string) {
	var (
		msg []byte
		err error
	)

	typingMsg := protocol.NewTypingEvent(value)
	if msg, err = json.Marshal(typingMsg); err != nil {
		logger.Error("[Client.MakeGuess] Failed to marshal TypingEvent: %v", err)
		os.Exit(1)
	}

	c.send(msg)
}

func (c *Client) LeaveMatch() {
	var (
		msg []byte
		err error
	)

	leaveMatchMsg := protocol.NewLeaveMatchEvent()
	if msg, err = json.Marshal(leaveMatchMsg); err != nil {
		logger.Error("[Client.MakeGuess] Failed to marshal LeaveMatchEvent: %v", err)
		os.Exit(1)
	}

	c.send(msg)
}

func (c *Client) send(payload []byte) {
	logger.Info("[Client.send] Sending event: %s", payload)
	if _, err := transport.SendEvent(c.Conn, payload); err != nil {
		logger.Error("[Client.send] Failed to send event: %v", err)
	}
}
