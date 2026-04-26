package client

import (
	"encoding/json"
	"fmt"
	"guessh/internal/logger"
	"guessh/internal/protocol"
	"guessh/internal/transport"
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

func (c *Client) CreateMatch(mode protocol.GameMode, format protocol.GameFormat, wordLen, rounds, turnTimeout int, playerName string) {
	createMatchEvent := protocol.NewCreateMatchEvent(mode, format, wordLen, rounds, turnTimeout, playerName)
	c.marshalAndSend(createMatchEvent)
}

func (c *Client) MakeGuess(guess string) {
	makeGuessEvent := protocol.NewMakeGuessEvent(guess)
	c.marshalAndSend(makeGuessEvent)
}

func (c *Client) JoinRoom(roomID string, playerName string) {
	joinRoomEvent := protocol.NewJoinRoomEvent(roomID, playerName)
	c.marshalAndSend(joinRoomEvent)
}

func (c *Client) Typing(value string) {
	typingEvent := protocol.NewTypingEvent(value)
	c.marshalAndSend(typingEvent)
}

func (c *Client) LeaveMatch() {
	leaveMatchEvent := protocol.NewLeaveMatchEvent()
	c.marshalAndSend(leaveMatchEvent)
}

func (c *Client) RequestRematch() {
	requestRematchEvent := protocol.NewRequestRematchEvent()
	c.marshalAndSend(requestRematchEvent)
}

func (c *Client) DenyRematch() {
	denyRematchEvent := protocol.NewDenyRematchEvent()
	c.marshalAndSend(denyRematchEvent)
}

func (c *Client) ReadyNextRound() {
	readyNextRoundEvent := protocol.NewReadyNextRoundEvent()
	c.marshalAndSend(readyNextRoundEvent)
}

func (c *Client) ShowStats() {
	showStatsEvent := protocol.NewShowStatsEvent()
	c.marshalAndSend(showStatsEvent)
}

func (c *Client) marshalAndSend(event any) {
	var (
		msg []byte
		err error
	)

	if msg, err = json.Marshal(event); err != nil {
		panic(fmt.Sprintf("[client] invariant violated: failed to marshal %T: %v", event, err))
	}

	c.Send(msg)
}

func (c *Client) Send(payload []byte) {
	logger.Info("[client] Sending event: %s", payload)
	if _, err := transport.SendEvent(c.Conn, payload); err != nil {
		logger.Error("[client] Failed to send event: %v", err)
	}
}
