package gnode

import (
	"fmt"
	"testing"
	"time"
)

type MsgSink struct {
	t *testing.T
}

func (sink *MsgSink) OnReceive(ctx *ProcessContext) {
	switch ctx.Msg().(type) {
	case *msgSysProcessStart:
		fmt.Printf("start: %v, %v\n", ctx.PID(), ctx.CoID())
		ctx.p.stop()
	case *msgSysProcessStop:
		fmt.Printf("stop: %v, %v\n", ctx.PID(), ctx.CoID())
	}
}

func TestProcess(t *testing.T) {
	p := newProcess(1, &MsgSink{t})
	p.start()

	<-time.After(5 * time.Second)
}
