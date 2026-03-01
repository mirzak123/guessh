package main

import (
	"fmt"
	"guessh/internal/logger"
	"guessh/internal/screen"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	wtea "github.com/charmbracelet/wish/bubbletea"
)

func main() {
	var (
		logFileName = "cli.log"
		logFile     *os.File
		err         error
	)

	if logFile, err = os.OpenFile("cli.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o777); err != nil {
		log.Fatalf("Failed to open log file %s: %v", logFileName, err)
	}

	logger.Init(logFile, logger.GetLogLevelFromEnv())
	var (
		server *ssh.Server
		addr   = "0.0.0.0:2222"
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

				return model, nil
			}),
		),
	)
	if err != nil {
		fmt.Printf("Could not create server: %v", err)
	}

	fmt.Printf("Listening on [%s]...", addr)
	if err = server.ListenAndServe(); err != nil {
		fmt.Printf("Failed starting server: %v", err)
	}

}
