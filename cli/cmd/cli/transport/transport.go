package transport

import (
	"log"
	"net"

	tea "github.com/charmbracelet/bubbletea"
)

type EventMsg string

func ListenForActivity(conn net.Conn, msg chan EventMsg) tea.Cmd {
	return func() tea.Msg {
		for {
			buffer := make([]byte, 1024)
			if _, err := conn.Read(buffer); err != nil {
				log.Printf("ListenForActivity error: %v", err)
				return err
			}
			msg <- EventMsg(buffer)
		}
	}
}

func WaitForEvent(msg chan EventMsg) tea.Cmd {
	return func() tea.Msg {
		return EventMsg(<-msg)
	}
}

func SendMessage(conn net.Conn, msg EventMsg) {
	// TODO: handle error
	conn.Write([]byte(msg))
}
