package gnode

import (
	"errors"
)

type ProcessContext struct {
	p   *Process
	c   *coroutine
	msg interface{}
}

func (ctx *ProcessContext) Msg() interface{} {
	return ctx.msg
}

func (ctx *ProcessContext) PID() uint32 {
	return ctx.p.id
}

func (ctx *ProcessContext) CoID() uint32 {
	return ctx.c.id
}

func (ctx *ProcessContext) SetResumeHandler(handler ResumeHandler) {
	ctx.p.setResumeHandler(handler)
}

func (ctx *ProcessContext) Call(to uint32, methodName string, request interface{},
	timeout uint32 /*millisecond*/) (interface{}, error) {
	msg := &msgSysCallRequest{
		callID:     ctx.p.nextCallID(),
		from:       ctx.p.id,
		to:         to,
		methodName: methodName,
		request:    request,
	}

	SendMessageTo(msg.to, msg)

	ret := ctx.c.yield(timeout)
	switch ret.(type) {
	case *msgSysCallResponse:
		res := ret.(*msgSysCallResponse)
		return res.response, res.err
	case error:
		return nil, ret.(error)
	default:
		return nil, errors.New("Call return unknown type value")
	}
}
