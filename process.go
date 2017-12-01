package gnode

import (
	"errors"
	"time"
)

type ResumeHandler func(msg interface{}) uint32

type suspendCoroutine struct {
	co         *coroutine
	resumeTime time.Time
}

type Process struct {
	id            uint32
	supCoroutines []*suspendCoroutine
	msgQ          *concurrentQueue
	rh            ResumeHandler
	lastCoID      uint32
	lastCallID    uint32
	frameTime     time.Time
	stopFlag      bool
	r             Reactor
}

func newProcess(id uint32, r Reactor) *Process {
	return &Process{
		id:            id,
		supCoroutines: make([]*suspendCoroutine, 0, 8),
		msgQ:          newConcurrentQueue(),
		rh:            nil,
		lastCoID:      0,
		lastCallID:    0,
		frameTime:     time.Time{},
		stopFlag:      false,
		r:             r,
	}
}

func (p *Process) start() {
	p.sendMessage(&msgSysProcessStart{})
	go p.loop()
}

func (p *Process) stop() {
	p.stopFlag = true
}

func (p *Process) loop() {
	p.frameTime = time.Now()

	for !p.stopFlag {
		p.processMsg()
		p.processTick()
	}

	co := newCoroutine(p.nextCoID())
	ret := co.start(p.r.OnReceive, &ProcessContext{p, co, &msgSysProcessStop{}})
	if ret != nil {
		// TODO: log here
	}

	p.cleanup()
}

func (p *Process) processMsg() {
	const maxProcessMsgCount = 64
	for i := 0; i < maxProcessMsgCount; i++ {
	begin:
		msg := p.msgQ.pop()
		if msg == nil {
			return
		}

		// resume handler
		if p.rh != nil {
			if coid := p.rh(msg); coid != 0 {
				for i, sc := range p.supCoroutines {
					if sc.co.id == coid {
						// resume this coroutine
						if sc.co.resume(msg) == nil {
							// coroutine is termination, remove it
							p.supCoroutines = append(p.supCoroutines[:i], p.supCoroutines[i+1:]...)
						}
						goto begin
					}
				}
			}
		}

		// start a new coroutine
		co := newCoroutine(p.nextCoID())
		yieldValue := co.start(p.r.OnReceive, &ProcessContext{p, co, msg})
		if yieldValue != nil {
			resumeTime := p.frameTime.Add(yieldValue.(time.Duration) * time.Millisecond)
			// coroutine is suspend, add to suspend list
			p.supCoroutines = append(p.supCoroutines, &suspendCoroutine{co, resumeTime})
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

		p.frameTime.Add(costTime)
	} else {
		sleepTime := frameInterval - costTime
		<-time.After(sleepTime)

		p.frameTime.Add(frameInterval)
	}
}

var errCoResumeTimeout = errors.New("coroutine resume timeout")

func (p *Process) onTick() {
	// scan suspend coroutines
	for i := len(p.supCoroutines) - 1; i >= 0; i-- {
		sc := p.supCoroutines[i]
		if p.frameTime.After(sc.resumeTime) || p.frameTime.Equal(sc.resumeTime) {
			sc.co.resume(errCoResumeTimeout)
			p.supCoroutines = append(p.supCoroutines[:i], p.supCoroutines[i+1:]...)
		}
	}

	// design timer service
	// TODO
}

func (p *Process) nextCoID() uint32 {
	p.lastCoID++
	return p.lastCoID
}

func (p *Process) sendMessage(msg interface{}) {
	p.msgQ.push(msg)
}

func (p *Process) setResumeHandler(handler ResumeHandler) {
	p.rh = handler
}

func (p *Process) nextCallID() uint32 {
	p.lastCallID++
	return p.lastCallID
}

func (p *Process) cleanup() {
}
