package tcp

type ServerEventHandler interface {
	OnSessionEstablish(*Session)
	OnSessionBreak(*Session)
}

type SessionHandler interface {
	OnMessage(*Session, interface{})
}
