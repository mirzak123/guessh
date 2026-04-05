package main

import (
	"bufio"
	"fmt"
	"guessh/internal/logger"
	"guessh/internal/transport"
	"net"
	"os"
	"strings"
)

const PROMPT = "> "

func main() {
	logger.EnsureLoggerSetup("cli.log")

	var (
		conn net.Conn
		err  error
	)

	if conn, err = transport.Connect(); err != nil {
		fmt.Printf("Could not connect to server: %v", err)
		os.Exit(1)
	}

	go func() {
		for {
			buf, err := transport.ReadServerEvent(conn)
			if err != nil {
				fmt.Printf("Server disconnected: %v", err)
				os.Exit(1)
			}

			event := string(buf)
			fmt.Printf("\r%s%s\n", strings.Repeat(" ", len(PROMPT)), event)
			fmt.Print(PROMPT)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(PROMPT)
		scanner.Scan()

		bytes := scanner.Bytes()
		if len(bytes) == 0 {
			continue
		}

		_, _ = transport.SendEvent(conn, bytes)
	}
}
