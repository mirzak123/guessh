package transport

import (
	"encoding/binary"
	"io"
	"net"
)

const EVENT_LEN_BYTES = 4

var byteOrder = binary.BigEndian

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
