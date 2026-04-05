package transport

import (
	"encoding/binary"
	"guessh/internal/config"
	"io"
	"net"
)

const EVENT_LEN_BYTES = 4

var byteOrder = binary.BigEndian

func Connect() (net.Conn, error) {
	serverAddr := config.GetEnv("GAME_SERVER_ADDR", "localhost:2480")
	return net.Dial("tcp", serverAddr)
}

func ReadServerEvent(conn net.Conn) ([]byte, error) {
	var eventLen uint32
	if err := binary.Read(conn, byteOrder, &eventLen); err != nil {
		return nil, err
	}

	buf := make([]byte, eventLen)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func SendEvent(conn net.Conn, payload []byte) (int, error) {
	payloadLen := uint32(len(payload))
	prefix := make([]byte, EVENT_LEN_BYTES)
	byteOrder.PutUint32(prefix, payloadLen)

	return conn.Write(append(prefix, payload...))
}
