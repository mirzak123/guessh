package transport

import (
	"encoding/binary"
	"log"
	"net"

	tea "github.com/charmbracelet/bubbletea"
)

const MESSAGE_LEN_BYTES = 4

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

func SendMessage(conn net.Conn, payload []byte) (int, error) {
	payloadLen := uint32(len(payload))
	prefix := make([]byte, MESSAGE_LEN_BYTES)
	binary.BigEndian.PutUint32(prefix, payloadLen)

	return conn.Write(append(prefix, payload...))
}
