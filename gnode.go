package gnode

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

type processMap struct {
	sync.RWMutex
	processes map[uint32]*Process
}

type gSystem struct {
	id      uint32
	pm      *processMap
	ipcMsgQ *concurrentQueue
	lastPID uint32
	stopC   chan int32
}

type ipcMessage struct {
	to  uint32
	msg interface{}
}

var gsys *gSystem

// Start make gsystem startup
func Start(nid uint32) {
	gsys = &gSystem{
		id:      nid,
		pm:      &processMap{processes: make(map[uint32]*Process)},
		ipcMsgQ: newConcurrentQueue(),
		lastPID: 0,
		stopC:   make(chan int32),
	}

	go gsys.loop()
}

// Stop make the world stop
func Stop() {
	if gsys == nil {
		return
	}

	// stop all processes
	gsys.pm.RLock()
	for _, p := range gsys.pm.processes {
		p.stop()
	}
	gsys.pm.RUnlock()

	gsys.pm.Lock()
	gsys.pm.processes = make(map[uint32]*Process) // Don't set nil
	gsys.pm.Unlock()

	gsys.stopC <- 0
}

func (g *gSystem) loop() {
	const gSystemFrameInterval time.Duration = 100 * time.Millisecond
	t := time.Now()
	l := list.New()
	for {
		l.Init()
		l = g.ipcMsgQ.exchange(l)
		for e := l.Front(); e != nil; e = e.Next() {
			imsg := e.Value.(ipcMessage)
			p := g.getProcess(imsg.to)
			if p != nil {
				p.postMessage(imsg.msg)
			}
		}

		now := time.Now()
		costTime := now.Sub(t)
		if costTime > gSystemFrameInterval {
			// System is busy
		} else {
			sleepTime := gSystemFrameInterval - costTime
			select {
			case <-time.After(sleepTime):
			case <-g.stopC:
				return
			}
		}

		t = now
	}
}

func (g *gSystem) getProcess(pid uint32) *Process {
	g.pm.RLock()
	defer g.pm.RUnlock()

	if p, ok := g.pm.processes[pid]; ok {
		return p
	}

	return nil
}

func (g *gSystem) addProcess(p *Process) bool {
	g.pm.Lock()
	_, exist := g.pm.processes[p.id]
	if !exist {
		g.pm.processes[p.id] = p
	}
	g.pm.Unlock()
	return !exist
}

func (g *gSystem) removeProcess(pid uint32) {
	g.pm.Lock()
	defer g.pm.Unlock()

	delete(g.pm.processes, pid)
}

func (g *gSystem) nextPID() uint32 {
	return atomic.AddUint32(&g.lastPID, 1)
}

// CreateProcess create a new process
func CreateProcess(r Reactor) uint32 {
	if gsys == nil {
		return 0
	}
	pid := gsys.nextPID()
	p := newProcess(pid, r)
	if !gsys.addProcess(p) {
		return 0
	}
	p.start()
	return pid
}

// DestroyProcess destroy a process by pid
func DestroyProcess(pid uint32) {
	if gsys == nil {
		return
	}

	p := gsys.getProcess(pid)
	if p == nil {
		return
	}

	gsys.removeProcess(p.id)
	p.stop()
}

// RouteMessage route a messag to the target process
func RouteMessage(target uint32, msg interface{}) {
	if gsys == nil {
		return
	}
	gsys.ipcMsgQ.push(ipcMessage{target, msg})
}
