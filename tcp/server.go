package tcp

import "net"

type Server struct {
	listener net.Listener
	handler  ServerEventHandler
}

func NewServer(handler ServerEventHandler) *Server {
	return &Server{
		listener: nil,
		handler:  handler,
	}
}

func (s *Server) Start(address string) bool {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}

	s.listener = l
	go s.accept()
	return true
}

func (s *Server) Stop() {
	s.listener.Close()
}

func (s *Server) accept() {
	defer s.listener.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// TODO: how to distinguish whether listen socket closed or accept error
			return
		}

		session := newSession(conn)
		s.handler.OnSessionEstablish(session)
	}
}
