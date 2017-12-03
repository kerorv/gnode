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
	case *MsgProcessStart:
		fmt.Printf("start: %v\n", ctx.PID())
		ctx.p.stop()
	case *MsgProcessStop:
		fmt.Printf("stop: %v\n", ctx.PID())
	}
}

func TestProcess(t *testing.T) {
	p := newProcess(1, &MsgSink{t})
	p.start()

	<-time.After(5 * time.Second)
}
