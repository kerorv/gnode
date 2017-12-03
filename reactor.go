package gnode

// Reactor is a message processor of the binding process
type Reactor interface {
	OnReceive(*ProcessContext)
}
