package main

import (
	"guessh/internal/game"
	"guessh/internal/logger"
	"guessh/internal/screen"
	"log"
	"os"
	"strings"

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

	logger.Init(logFile, getLogLevelFromEnv())

	game.FiveLetterWords = game.ExtractWordsFromFile("/Users/mirza/code/personal/guessh/words/five-letter.txt")
	game.SixLetterWords = game.ExtractWordsFromFile("/Users/mirza/code/personal/guessh/words/six-letter.txt")
	game.SevenLetterWords = game.ExtractWordsFromFile("/Users/mirza/code/personal/guessh/words/seven-letter.txt")

	p := tea.NewProgram(
		screen.InitialModel(),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		logger.Error("Error running program: %v\n", err)
		os.Exit(1)
	}
}

func getLogLevelFromEnv() logger.LogLevel {
	logLevelStr, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		logLevelStr = "INFO"
	}

	logLevelStr = strings.ToUpper(logLevelStr)

	switch logLevelStr {
	case "DEBUG":
		return logger.LogLevelDebug
	default:
		return logger.LogLevelInfo
	}
}
