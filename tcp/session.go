package tcp

import (
	"net"

	"github.com/kerorv/gnode"
)

var (
	writeChanCap    = 64
	readBufferSize  = 1024
	writeBufferSize = 1024
)

type session struct {
	conn    net.Conn
	handler uint32
	codec   Codec
	writeC  chan interface{}
}

func newSession(conn net.Conn, backend uint32, codec Codec) *session {
	return &session{
		conn:    conn,
		handler: backend,
		codec:   codec,
		writeC:  make(chan interface{}, writeChanCap),
	}
}

func (s *session) start() {
	go s.readConn()
	go s.writeConn()
}

func (s *session) readConn() {
	var recvLen int
	var p = make(packet, readBufferSize)

	for {
		size, _ := s.conn.Read(p)
		recvLen = recvLen + size

		// parse buffer
		if recvLen < packetHeaderSize {
			return
		}

		for rpos := p.size(); rpos < len(p); {
			p1 := p[rpos:]
			if len(p1) < p1.size() {
				break
			}

			if msg, ok := s.codec.Decode(p1); ok {
				gnode.RouteMessage(s.handler, msg)
			} else {
				// TODO: break connection
				break
			}
		}
	}
}

func (s *session) writeConn() {
	for {
		select {
		case msg := <-s.writeC:
			if msg == nil {
				// TODO: close the connection
				// Let readConn do it?
				return
			}

			if p, ok := s.codec.Encode(msg); ok {
				var written int
				for count, err := s.conn.Write(p); written < len(p); written += count {
					if err != nil {
						// TODO: break the connection
						return
					}
				}
			}
		}
	}
}

func (s *session) stop() {
	s.conn.Close()
}
