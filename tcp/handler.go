package tcp

type NewSessionHandler interface {
	OnNewSession(*Session)
}

type SessionEventHandler interface {
	OnMessage(*Session, interface{})
	OnBroken(*Session)
}
