package main

import (
	"guessh/internal/logger"
	"guessh/internal/screen"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	logger.EnsureLoggerSetup("cli.log")

	p := tea.NewProgram(screen.InitialModel())
	if _, err := p.Run(); err != nil {
		logger.Error("Error running program: %v\n", err)
		os.Exit(1)
	}
}
