package main

import (
	"guessh/internal/logger"
	"guessh/internal/screen"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
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

	p := tea.NewProgram(screen.InitialModel())
	if _, err := p.Run(); err != nil {
		logger.Error("Error running program: %v\n", err)
		os.Exit(1)
	}
}
