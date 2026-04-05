package main

import (
	"bufio"
	"bytes"
	"fmt"
	"guessh/internal/client"
	"guessh/internal/logger"
	"guessh/internal/transport"
	"net"
	"os"
	"strings"
)

const PROMPT = "> "

var (
	quit  = []byte("quit")
	stats = []byte("stats")
)

func main() {
	logger.EnsureLoggerSetup("cli.log")

	var (
		conn net.Conn
		err  error
	)

	if conn, err = transport.Connect(); err != nil {
		fmt.Printf("Could not connect to server: %v\n", err)
		os.Exit(1)
	}

	go func() {
		for {
			buf, err := transport.ReadServerEvent(conn)
			if err != nil {
				fmt.Printf("\nServer disconnected: %v\n", err)
				os.Exit(1)
			}

			event := string(buf)
			fmt.Printf("\r%s%s\n", strings.Repeat(" ", len(PROMPT)), event)
			fmt.Print(PROMPT)
		}
	}()

	cl := client.NewClient(conn)
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(PROMPT)
		scanner.Scan()

		bs := scanner.Bytes()
		if len(bs) == 0 {
			continue
		}

		if bytes.Equal(quit, bs) {
			os.Exit(0)
		}

		if bytes.Equal(stats, bs) {
			cl.ShowStats()
			continue
		}

		cl.Send(bs)
	}
}
