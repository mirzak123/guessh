package game

import (
	"guessh/internal/config"
	"guessh/internal/logger"
	"os"
	"path/filepath"
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
		basePath := config.GetEnv("WORDS_PATH", "./words")

		FiveLetterWords = ExtractWordsFromFile(filepath.Join(basePath, "five-letter.txt"))
		SixLetterWords = ExtractWordsFromFile(filepath.Join(basePath, "six-letter.txt"))
		SevenLetterWords = ExtractWordsFromFile(filepath.Join(basePath, "seven-letter.txt"))
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
