package connection

import (
	"log"
	"net"

	tea "github.com/charmbracelet/bubbletea"
)

type EventMsg string

func ListenForActivity(conn net.Conn, sub chan string) tea.Cmd {
	return func() tea.Msg {
		for {
			buffer := make([]byte, 1024)
			if _, err := conn.Read(buffer); err != nil {
				log.Printf("ListenForActivity error: %v", err)
				return err
			}
			sub <- string(buffer)
		}
	}
}

func WaitForEvent(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return EventMsg(<-sub)
	}
}

func SendMessage(conn net.Conn, msg string) {
	// TODO: handle error
	conn.Write([]byte(msg))
}
