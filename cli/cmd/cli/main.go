package main

import (
	"guessh/cmd/cli/screen"
	"log"
)

func main() {
	err := screen.StartMenu.Run()
	if err != nil {
		log.Fatal(err)
	}
}
