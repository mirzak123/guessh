package transport

import (
	"encoding/binary"
	"guessh/internal/logger"
	"io"
	"net"

	tea "github.com/charmbracelet/bubbletea"
)

const EVENT_LEN_BYTES = 4

type EventMsg string

type ServerDisconnectedMsg struct {
	Err error
}

func ListenForActivity(conn net.Conn, msg chan EventMsg) tea.Cmd {
	return func() tea.Msg {
		for {
			var length uint32
			if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
				logger.Error("[ListenForActivity] error: %v", err)
				return ServerDisconnectedMsg{
					Err: err,
				}
			}

			buffer := make([]byte, length)

			if _, err := io.ReadFull(conn, buffer); err != nil {
				logger.Error("[ListenForActivity] error: %v", err)
				return ServerDisconnectedMsg{
					Err: err,
				}
			}

			logger.Debug("Received network data: %s", buffer)
			msg <- EventMsg(buffer)
		}
	}
}

func WaitForEvent(msg chan EventMsg) tea.Cmd {
	return func() tea.Msg {
		return EventMsg(<-msg)
	}
}

func SendEvent(conn net.Conn, payload []byte) (int, error) {
	payloadLen := uint32(len(payload))
	prefix := make([]byte, EVENT_LEN_BYTES)
	binary.BigEndian.PutUint32(prefix, payloadLen)

	return conn.Write(append(prefix, payload...))
}
