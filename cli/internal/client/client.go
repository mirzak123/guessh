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

	createMatchEvent := protocol.NewCreateMatchEvent(mode, wordLen, rounds, playerName)
	if msg, err = json.Marshal(createMatchEvent); err != nil {
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

	makeGuessEvent := protocol.NewMakeGuessEvent(guess)
	if msg, err = json.Marshal(makeGuessEvent); err != nil {
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

	joinRoomEvent := protocol.NewJoinRoomEvent(roomID, playerName)
	if msg, err = json.Marshal(joinRoomEvent); err != nil {
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

	typingEvent := protocol.NewTypingEvent(value)
	if msg, err = json.Marshal(typingEvent); err != nil {
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

	leaveMatchEvent := protocol.NewLeaveMatchEvent()
	if msg, err = json.Marshal(leaveMatchEvent); err != nil {
		logger.Error("[Client.MakeGuess] Failed to marshal LeaveMatchEvent: %v", err)
		os.Exit(1)
	}

	c.send(msg)
}

func (c *Client) RequestRematch() {
	var (
		msg []byte
		err error
	)

	requestRematchEvent := protocol.NewRequestRematchEvent()
	if msg, err = json.Marshal(requestRematchEvent); err != nil {
		logger.Error("[Client.MakeGuess] Failed to marshal RequestRematchEvent: %v", err)
		os.Exit(1)
	}

	c.send(msg)
}

func (c *Client) DenyRematch() {
	var (
		msg []byte
		err error
	)

	denyRematchEvent := protocol.NewDenyRematchEvent()
	if msg, err = json.Marshal(denyRematchEvent); err != nil {
		logger.Error("[Client.MakeGuess] Failed to marshal DenyRematchEvent: %v", err)
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
