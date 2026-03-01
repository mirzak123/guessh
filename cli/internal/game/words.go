package game

import (
	"guessh/internal/logger"
	"os"
	"strings"
	"sync"
)

var (
	FiveLetterWords  []string
	SixLetterWords   []string
	SevenLetterWords []string

	loadOnce sync.Once
)

func EnsureDictionariesLoaded() {
	loadOnce.Do(func() {
		// This executes exactly once even if CLI is served through SSH and multiple clients connect
		FiveLetterWords = ExtractWordsFromFile("/Users/mirza/code/personal/guessh/words/five-letter.txt")
		SixLetterWords = ExtractWordsFromFile("/Users/mirza/code/personal/guessh/words/six-letter.txt")
		SevenLetterWords = ExtractWordsFromFile("/Users/mirza/code/personal/guessh/words/seven-letter.txt")
	})
}

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
