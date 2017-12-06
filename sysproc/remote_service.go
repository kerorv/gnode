package sysproc

import (
	"github.com/kerorv/gnode"
	"github.com/kerorv/gnode/tcp"
)

type RemoteService struct {
	server      *tcp.Server
	sessions    map[uint32]*tcp.Session
	BindAddress string
}

func (rs *RemoteService) OnReceive(ctx *gnode.ProcessContext) {
	switch ctx.Msg().(type) {
	case *gnode.MsgProcessStart:
		if !rs.init() {
			gnode.DestroyProcess(ctx.PID())
		}
	case *gnode.MsgProcessStop:
		rs.release()
	}
}

func (rs *RemoteService) init() bool {
	rs.server = tcp.NewServer(rs)
	return rs.server.Start(rs.BindAddress)
}

func (rs *RemoteService) release() {
	rs.server.Stop()
}

func (rs *RemoteService) OnNewSession(session *tcp.Session) {
	var id uint32
	rs.sessions[id] = session
	session.Start(rs, new(ProtoCodec))
}

func (rs *RemoteService) OnMessage(session *Session, msg interface{}) {

}

func (rs *RemoteService) OnBreak(session *Session) {

}
