package gnode

import (
	"fmt"
	"reflect"
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

type Adder struct {
}

func (self *Adder) Add(n1 int, n2 int) int {
	return n1 + n2
}

func (self *Adder) onCall(ctx *ProcessContext, callid uint32, from uint32, methodName string, params interface{}) {
	method := reflect.ValueOf(self).MethodByName(methodName)
	inputs := make([]reflect.Value, 1)
	inputs[0] = reflect.ValueOf(params)
	ret := method.Call(inputs)

	response := &msgSysCallResponse{callid, ctx.PID(), from, ret, nil}
	SendMsgTo(from, response)
}

func (self *Adder) OnReceive(ctx *ProcessContext) {
	switch ctx.Msg().(type) {
	case *msgSysCallRequest:
		call := ctx.Msg().(*msgSysCallRequest)
		self.onCall(ctx, call.callID, call.from, call.methodName, call.request)
	}
}

func TestProcess(t *testing.T) {
	p := newProcess(1, &MsgSink{t})
	p.start()

	<-time.After(5 * time.Second)
}

func TestProcess2(t *testing.T) {
	p1 := newProcess(1, &MsgSink{t})
	p1.start()

	p2 := newProcess(2, &MsgSink{t})
	p2.start()

	<-time.After(5 * time.Second)
}
