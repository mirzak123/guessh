package main

import (
	"guessh/internal/screen"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	log.SetFlags(log.Lshortfile)
	if _, err := tea.LogToFile("cli.log", ""); err != nil {
		log.Fatalf("tea.LogToFile failed: %v", err)
	}

	p := tea.NewProgram(
		screen.InitialModel(),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v\n", err)
	}
}
