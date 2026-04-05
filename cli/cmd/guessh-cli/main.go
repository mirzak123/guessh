package main

import "guessh/internal/logger"

func main() {
	logger.EnsureLoggerSetup("cli.log")

	logger.Info("hello world")
}
