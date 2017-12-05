package gnode

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

type suspendCoroutine struct {
	co         *coroutine
	resumeTime time.Time
	callID     uint32
}

type yieldValue struct {
	timout time.Duration
	callID uint32
}

type resumeValue struct {
	response []interface{}
	err      error
}

// Process is an execute unit in gnode system.
type Process struct {
	id            uint32
	supCoroutines []*suspendCoroutine
	msgQ          *concurrentQueue
	lastCoID      uint32
	lastCallID    uint32
	frameTime     time.Time
	stopFlag      bool
	stopC         chan int32
	doneWG        sync.WaitGroup
	r             Reactor
}

func newProcess(id uint32, r Reactor) *Process {
	return &Process{
		id:            id,
		supCoroutines: make([]*suspendCoroutine, 0, 8),
		msgQ:          newConcurrentQueue(),
		frameTime:     time.Time{},
		stopFlag:      false,
		stopC:         make(chan int32),
		r:             r,
	}
}

func (p *Process) start() {
	p.postMessage(&MsgProcessStart{})
	go p.loop()
}

func (p *Process) stop() {
	p.stopC <- 0
	p.doneWG.Wait()
}

func (p *Process) internalStop() {
	p.stopFlag = true
}

func (p *Process) loop() {
	p.frameTime = time.Now()
	p.doneWG.Add(1)

	for !p.stopFlag {
		p.processMsg()
		p.processTick()
	}

	p.invokeReactorDirectly(&ProcessContext{p, nil, &MsgProcessStop{}})
	p.cleanup()

	p.doneWG.Done()
}

func (p *Process) processMsg() {
	const maxProcessMsgCount = 64
	for i := 0; i < maxProcessMsgCount; i++ {
		msg := p.msgQ.pop()
		if msg == nil {
			return
		}

		// handle call
		switch msg.(type) {
		case *msgProcessCallRequest:
			callReq := msg.(*msgProcessCallRequest)
			p.onCallRequest(callReq.from, callReq.callID, callReq.methodName, callReq.request)
			continue
		case *msgProcessCallResponse:
			callRep := msg.(*msgProcessCallResponse)
			p.onCallResponse(callRep.from, callRep.callID, callRep.response, callRep.err)
			continue
		}

		// start a new coroutine
		co := newCoroutine(p.nextCoID())
		yd := co.start(p.r.OnReceive, &ProcessContext{p, co, msg})
		p.onYield(co, yd)
	}
}

func (p *Process) processTick() {
	p.onTick()

	// TODO: use monotonic time
	const frameInterval time.Duration = 100 * time.Millisecond
	now := time.Now()
	costTime := now.Sub(p.frameTime)
	if costTime > frameInterval {
		// process is busy
		// TODO: log here

		select {
		case <-p.stopC:
			p.stopFlag = true
		default:
		}
		p.frameTime.Add(costTime)
	} else {
		sleepTime := frameInterval - costTime
		select {
		case <-p.stopC:
			p.stopFlag = true
		case <-time.After(sleepTime):
		}
		p.frameTime.Add(frameInterval)
	}
}

var errCoResumeTimeout = errors.New("coroutine resume timeout")

func (p *Process) onTick() {
	// scan suspend coroutines
	colen := len(p.supCoroutines) // MUST use this len cache since onYield() would append coroutine
	for i := colen - 1; i >= 0; i-- {
		sc := p.supCoroutines[i]
		if p.frameTime.After(sc.resumeTime) || p.frameTime.Equal(sc.resumeTime) {
			// remove this suspend coroutine
			p.supCoroutines = append(p.supCoroutines[:i], p.supCoroutines[i+1:]...)
			// resume coroutine
			yd := sc.co.resume(&resumeValue{nil, errCoResumeTimeout})
			p.onYield(sc.co, yd)
		}
	}

	// design timer service
	// TODO
}

func (p *Process) onCallRequest(from uint32, callID uint32, methodName string, params []interface{}) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("rpc call error: %v", r)
			}

			response := &msgProcessCallResponse{callID, p.id, from, nil, err}
			RouteMessage(from, response)
		}
	}()

	method := reflect.ValueOf(p.r).MethodByName(methodName)
	inputs := make([]reflect.Value, len(params))
	for i, p := range params {
		inputs[i] = reflect.ValueOf(p)
	}
	outputs := method.Call(inputs)
	ret := make([]interface{}, len(outputs))
	for i, o := range outputs {
		ret[i] = o.Interface()
	}

	response := &msgProcessCallResponse{callID, p.id, from, ret, nil}
	RouteMessage(from, response)
}

func (p *Process) onCallResponse(from uint32, callID uint32, response []interface{}, err error) {
	for i, supCo := range p.supCoroutines {
		if supCo.callID == callID {
			// remove this suspend coroutine
			p.supCoroutines = append(p.supCoroutines[:i], p.supCoroutines[i+1:]...)
			// resume coroutine
			yd := supCo.co.resume(&resumeValue{response, err})
			p.onYield(supCo.co, yd)
			return
		}
	}
}

func (p *Process) onYield(co *coroutine, ydata interface{}) {
	if ydata != nil {
		// coroutine suspend
		yv := ydata.(*yieldValue)
		resumeTime := p.frameTime.Add(yv.timout)
		// add to suspend list
		p.supCoroutines = append(p.supCoroutines, &suspendCoroutine{co, resumeTime, yv.callID})
	} else {
		// coroutine is finished
		if co.panicE != nil {
			p.invokeReactorDirectly(&ProcessContext{p, nil, &MsgProcessCoroutinePanic{co.panicE}})
		}
	}
}

func (p *Process) invokeReactorDirectly(ctx *ProcessContext) {
	defer func() {
		if r := recover(); r != nil {
			// TODO: log here
		}
	}()

	p.r.OnReceive(ctx)
}

func (p *Process) nextCoID() uint32 {
	p.lastCoID++
	return p.lastCoID
}

func (p *Process) postMessage(msg interface{}) {
	p.msgQ.push(msg)
}

func (p *Process) nextCallID() uint32 {
	p.lastCallID++
	return p.lastCallID
}

func (p *Process) cleanup() {
}
