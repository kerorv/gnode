package tcp

import "net"

type Server struct {
	l        net.Listener
	handler  uint32
	codec    Codec
	sessions []*session
}

func NewServer(backend uint32, codec Codec) *Server {
	return &Server{
		l:        nil,
		handler:  backend,
		codec:    codec,
		sessions: make([]*session, 0, 64),
	}
}

func (s *Server) Start(address string) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	s.l = l
	go s.accept()
	return nil
}

func (s *Server) accept() {
	defer s.l.Close()

	for {
		conn, err := s.l.Accept()
		if err != nil {
			// TODO: how to distinguish whether listen socket closed or accept error
			return
		}

		ss := newSession(conn, s.handler, s.codec)
		s.sessions = append(s.sessions, ss)
		ss.start()
	}
}

func (s *Server) Stop() {
	s.l.Close()

	for _, ss := range s.sessions {
		ss.stop()
	}
}
