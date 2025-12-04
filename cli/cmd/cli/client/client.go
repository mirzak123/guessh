package client

import "net"

type Client struct {
	Conn net.Conn
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		Conn: conn,
	}
}

func (c *Client) CreateMatch() {

}

func (c *Client) MakeGuess(guess string) {

}
