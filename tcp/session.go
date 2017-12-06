package tcp

import "net"

var (
	writeChanCap  = 64
	readBufferCap = 1024
)

type Session struct {
	conn    net.Conn
	handler SessionEventHandler
	codec   Codec
	writeC  chan interface{}
}

func newSession(conn net.Conn) *Session {
	return &Session{
		conn:    conn,
		handler: nil,
		codec:   nil,
		writeC:  make(chan interface{}, writeChanCap),
	}
}

func (s *Session) Start(handler SessionEventHandler, codec Codec) {
	s.handler = handler
	s.codec = codec

	go s.readConn()
	go s.writeConn()
}

func (s *Session) Write(msg interface{}) {
	s.writeC <- msg
}

func (s *Session) PeerAddress() string {
	return s.conn.RemoteAddr().String()
}

func (s *Session) Stop() {
	s.conn.Close()
}

func (s *Session) readConn() {
	var recvLen int
	var p = make(packet, 0, readBufferCap)

	for {
		size, _ := s.conn.Read(p)
		recvLen = recvLen + size

		// parse buffer
		if recvLen < packetHeaderSize {
			break
		}

		for rpos := p.size(); rpos < len(p); {
			p1 := p[rpos:]
			if len(p1) < p1.size() {
				break
			}

			if msg, ok := s.codec.Decode(p1); ok {
				s.handler.OnMessage(s, msg)
			} else {
				// TODO: break connection
				break
			}
		}
	}

	s.handler.OnBreak(s)
}

func (s *Session) writeConn() {
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
