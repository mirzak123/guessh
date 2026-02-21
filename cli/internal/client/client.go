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

func (c *Client) CreateMatch(mode protocol.GameMode, wordLen, rounds int) {
	var (
		msg []byte
		err error
	)

	createMatchMsg := protocol.NewCreateMatchMessage(mode, wordLen, rounds)
	if msg, err = json.Marshal(createMatchMsg); err != nil {
		logger.Error("[Client.CreateMatch] Failed to marshal CreateMatchMessage: %v", err)
		os.Exit(1)
	}

	c.send(msg)
}

func (c *Client) MakeGuess(guess string) {
	var (
		msg []byte
		err error
	)

	makeGuessMsg := protocol.NewMakeGuessMessage(guess)
	if msg, err = json.Marshal(makeGuessMsg); err != nil {
		logger.Error("[Client.MakeGuess] Failed to marshal CreateMatchMessage: %v", err)
		os.Exit(1)
	}

	c.send(msg)
}

func (c *Client) JoinRoom(roomID string) {
	var (
		msg []byte
		err error
	)

	joinRoomMsg := protocol.NewJoinRoomMessage(roomID)
	if msg, err = json.Marshal(joinRoomMsg); err != nil {
		logger.Error("[Client.MakeGuess] Failed to marshal JoinRoomMessage: %v", err)
		os.Exit(1)
	}

	c.send(msg)
}

func (c *Client) Typing(value string) {
	var (
		msg []byte
		err error
	)

	typingMsg := protocol.NewTypingMessage(value)
	if msg, err = json.Marshal(typingMsg); err != nil {
		logger.Error("[Client.MakeGuess] Failed to marshal TypingMessage: %v", err)
		os.Exit(1)
	}

	c.send(msg)
}

func (c *Client) send(payload []byte) {
	logger.Info("[Client.send] Sending message: %s", payload)
	if _, err := transport.SendMessage(c.Conn, payload); err != nil {
		logger.Error("[Client.send] Failed to send message: %v", err)
	}
}
