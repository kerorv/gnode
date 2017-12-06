package tcp

import (
	"net"
	"time"
)

func Connect(address string) *Session {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil
	}

	return newSession(conn)
}

func ConnectTimeout(address string, timeout time.Duration) *Session {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil
	}

	return newSession(conn)
}
