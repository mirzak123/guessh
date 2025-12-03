package transport

import (
	"log"
	"net"

	tea "github.com/charmbracelet/bubbletea"
)

type EventMsg string

func ListenForActivity(conn net.Conn, sub chan EventMsg) tea.Cmd {
	return func() tea.Msg {
		for {
			buffer := make([]byte, 1024)
			if _, err := conn.Read(buffer); err != nil {
				log.Printf("ListenForActivity error: %v", err)
				return err
			}
			sub <- EventMsg(string(buffer))
		}
	}
}

func WaitForEvent(sub chan EventMsg) tea.Cmd {
	return func() tea.Msg {
		return EventMsg(<-sub)
	}
}

func SendMessage(conn net.Conn, msg string) {
	// TODO: handle error
	conn.Write([]byte(msg))
}
