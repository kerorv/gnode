package gnode

import (
	"errors"
	"time"
)

// ProcessContext is context of a process
type ProcessContext struct {
	p   *Process
	c   *coroutine // may be nil
	msg interface{}
}

// Msg return current received message
func (ctx *ProcessContext) Msg() interface{} {
	return ctx.msg
}

// PID return id of current process
func (ctx *ProcessContext) PID() uint32 {
	return ctx.p.id
}

// PostMessage post a message to the process
func (ctx *ProcessContext) PostMessage(to uint32, msg interface{}) {
	if to == ctx.p.id {
		ctx.p.postMessage(msg)
	} else {
		RouteMessage(to, msg)
	}
}

var errCallUnknowReturn = errors.New("Call return unknown type value")
var errCallNoCoroutine = errors.New("Call fail since hasn't coroutine")
var errCallTheSameProcess = errors.New("Call fail since caller and callee are in the same process")

// NewRPC create a new rpc
func (ctx *ProcessContext) NewRPC(to uint32, methodName string, timeout uint32) *RPC {
	return &RPC{ctx.p, ctx.c, to, methodName, time.Duration(timeout) * time.Millisecond}
}

// RPC present a call struct
type RPC struct {
	p       *Process
	c       *coroutine
	to      uint32
	method  string
	timeout time.Duration
}

// Call invoke remote method
func (rpc *RPC) Call(params ...interface{}) ([]interface{}, error) {
	if rpc.c == nil {
		// coroutine may be nil when process is stopping
		return nil, errCallNoCoroutine
	}

	if rpc.to == rpc.p.id {
		return nil, errCallTheSameProcess
	}

	msg := &msgProcessCallRequest{
		callID:     rpc.p.nextCallID(),
		from:       rpc.p.id,
		to:         rpc.to,
		methodName: rpc.method,
		request:    params,
	}

	RouteMessage(rpc.to, msg)
	yv := &yieldValue{rpc.timeout, msg.callID}

	ret := rpc.c.yield(yv)
	switch ret.(type) {
	case *resumeValue:
		rv := ret.(*resumeValue)
		return rv.response, rv.err
	default:
		return nil, errCallUnknowReturn
	}
}
