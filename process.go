package gnode

import (
	"errors"
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
	timout uint32
	callID uint32
}

type resumeValue struct {
	response interface{}
	err      error
}

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
	p.sendMessage(&MsgProcessStart{})
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

	p.r.OnReceive(&ProcessContext{p, nil, &MsgProcessStop{}})
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
			p.onCallReqest(callReq.from, callReq.callID, callReq.methodName, callReq.request)
			continue
		case *msgProcessCallResponse:
			callRep := msg.(*msgProcessCallResponse)
			p.onCallResponse(callRep.from, callRep.callID, callRep.response, callRep.err)
			continue
		}

		// start a new coroutine
		co := newCoroutine(p.nextCoID())
		yieldData := co.start(p.r.OnReceive, &ProcessContext{p, co, msg})
		if yieldData != nil {
			yv := yieldData.(*yieldValue)
			resumeTime := p.frameTime.Add(time.Duration(yv.timout) * time.Millisecond)
			// coroutine is suspend, add to suspend list
			p.supCoroutines = append(p.supCoroutines, &suspendCoroutine{co, resumeTime, yv.callID})
		}
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
	for i := len(p.supCoroutines) - 1; i >= 0; i-- {
		sc := p.supCoroutines[i]
		if p.frameTime.After(sc.resumeTime) || p.frameTime.Equal(sc.resumeTime) {
			sc.co.resume(&resumeValue{nil, errCoResumeTimeout})
			// remove this coroutine
			p.supCoroutines = append(p.supCoroutines[:i], p.supCoroutines[i+1:]...)
		}
	}

	// design timer service
	// TODO
}

func (p *Process) onCallReqest(from uint32, callID uint32, methodName string, param interface{}) {
	method := reflect.ValueOf(p.r).MethodByName(methodName)
	inputs := make([]reflect.Value, 1)
	inputs[0] = reflect.ValueOf(param)
	ret := method.Call(inputs)

	response := &msgProcessCallResponse{callID, p.id, from, ret, nil}
	SendMessageTo(from, response)
}

func (p *Process) onCallResponse(from uint32, callID uint32, response interface{}, err error) {
	for i, supCo := range p.supCoroutines {
		if supCo.callID == callID {
			p.supCoroutines = append(p.supCoroutines[:i], p.supCoroutines[i+1:]...)
			rv := &resumeValue{response, err}
			supCo.co.resume(rv)
			return
		}
	}
}

func (p *Process) nextCoID() uint32 {
	p.lastCoID++
	return p.lastCoID
}

func (p *Process) sendMessage(msg interface{}) {
	p.msgQ.push(msg)
}

func (p *Process) nextCallID() uint32 {
	p.lastCallID++
	return p.lastCallID
}

func (p *Process) cleanup() {
}
