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
	close(s.writeC)
	s.conn.Close()
}

func (s *Session) readConn() {
	var p = make(Packet, 0, readBufferCap)

	for {
		_, err := s.conn.Read(p)
		if err != nil {
			s.conn.Close()
			break
		}

		// parse buffer
		if len(p) < PacketHeaderSize {
			continue
		}

		if int(p.Size()) > cap(p) {
			p1 := make(Packet, 0, p.Size())
			p1 = append(p1, p...)
			p = p1
		}

		var rpos uint16
		for int(rpos) < len(p) {
			sub := p[rpos:]

			if len(sub) < int(sub.Size()) {
				break
			}

			if msg, ok := s.codec.Decode(sub); ok {
				s.handler.OnMessage(s, msg)
				rpos += sub.Size()
			} else {
				// TODO: break connection
				goto end
			}
		}

		if rpos > 0 {
			p = append(p, p[rpos:]...)
		}
	}

end:
	s.handler.OnBroken(s)
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
