package game

import (
	"guessh/internal/logger"
	"os"
	"strings"
)

var (
	FiveLetterWords  []string
	SixLetterWords   []string
	SevenLetterWords []string
)

func ExtractWordsFromFile(path string) []string {
	var (
		content []byte
		err     error
	)

	if content, err = os.ReadFile(path); err != nil {
		logger.Error("Could not open word file %s: %v", path, err)
		os.Exit(1)
	}

	words := strings.Split(string(content), "\n")

	return words
}
