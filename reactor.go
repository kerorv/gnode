package gnode

type Reactor interface {
	OnReceive(*ProcessContext)
}
