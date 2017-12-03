package gnode

import (
	"errors"
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

var errCallUnknowReturn = errors.New("Call return unknown type value")
var errCallNoCoroutine = errors.New("Call fail since hasn't coroutine")

// Call is a remote function call, it will block current coroutine,
// but won't block process.
func (ctx *ProcessContext) Call(to uint32, methodName string, request interface{},
	timeout uint32 /*millisecond*/) (interface{}, error) {
	if ctx.c == nil {
		// coroutine may be nil when process is stopping
		return nil, errCallNoCoroutine
	}

	msg := &msgProcessCallRequest{
		callID:     ctx.p.nextCallID(),
		from:       ctx.p.id,
		to:         to,
		methodName: methodName,
		request:    request,
	}

	SendMessageTo(msg.to, msg)
	yv := &yieldValue{timeout, msg.callID}

	ret := ctx.c.yield(yv)
	switch ret.(type) {
	case *resumeValue:
		rv := ret.(*resumeValue)
		return rv.response, rv.err
	default:
		return nil, errCallUnknowReturn
	}
}
