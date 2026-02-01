package client

import (
	"encoding/json"
	"guessh/internal/protocol"
	"guessh/internal/transport"
	"log"
	"net"
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
		log.Fatalf("[Client.CreateMatch] Failed to marshal CreateMatchMessage: %v", err)
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
		log.Fatalf("[Client.MakeGuess] Failed to marshal CreateMatchMessage: %v", err)
	}

	c.send(msg)
}

func (c *Client) send(payload []byte) {
	log.Printf("[Client.send] Sending message: %s", payload)
	if _, err := transport.SendMessage(c.Conn, payload); err != nil {
		log.Printf("[Client.send] Failed to send message: %v", err)
	}
}
