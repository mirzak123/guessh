package transport

import (
	"guessh/internal/logger"
	"net"

	tea "github.com/charmbracelet/bubbletea"
)

type EventMsg string

type ServerDisconnectedMsg struct {
	Err error
}

func ListenForActivity(conn net.Conn, msg chan EventMsg) tea.Cmd {
	return func() tea.Msg {
		for {
			buffer, err := ReadServerEvent(conn)
			if err != nil {
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
