package main

import (
	"guessh/internal/screen"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	tea.LogToFile("cli.log", "")

	p := tea.NewProgram(
		screen.InitialModel(),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v\n", err)
	}
}
