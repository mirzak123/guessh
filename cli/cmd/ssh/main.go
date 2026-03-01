package main

import (
	"guessh/internal/logger"
	"guessh/internal/screen"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	wtea "github.com/charmbracelet/wish/bubbletea"
)

func main() {
	logger.EnsureLoggerSetup("ssh.log")

	var (
		server *ssh.Server
		addr   = "0.0.0.0:2222"
		err    error
	)

	server, err = wish.NewServer(
		wish.WithAddress(addr),
		wish.WithIdleTimeout(10*time.Minute),
		wish.WithMiddleware(
			wtea.Middleware(func(session ssh.Session) (tea.Model, []tea.ProgramOption) {
				_, _, active := session.Pty()
				if !active {
					wish.Fatalf(session, "no active terminal, skipping")
				}
				model := screen.InitialModel()
				model.SetSSHContext(session.Context())

				go func() {
					<-session.Context().Done()
					if model.GetClient() != nil && model.GetClient().Conn != nil {
						if err := model.GetClient().Conn.Close(); err != nil {
							logger.Error("Failed to close TCP connection for %s: %v", session.User(), err)
							os.Exit(1)
						}
						logger.Info("SSH session closed: TCP connection for %s forced shut", session.User())
					}
				}()

				return model, nil
			}),
		),
	)
	if err != nil {
		logger.Error("Could not create server: %v", err)
		os.Exit(1)
	}

	logger.Info("Listening on [%s]...", addr)
	if err = server.ListenAndServe(); err != nil {
		logger.Error("Failed starting server: %v", err)
		os.Exit(1)
	}

}
