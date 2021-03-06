package gnode

import "fmt"

const (
	suspend = 1
	running = 2
	stop    = 3
)

type coroutine struct {
	id      uint32
	status  int32
	panicE  error
	yieldC  chan interface{}
	resumeC chan interface{}
}

type coroutineFunc func(*ProcessContext)

func newCoroutine(id uint32) *coroutine {
	return &coroutine{
		id:      id,
		status:  stop,
		panicE:  nil,
		yieldC:  make(chan interface{}),
		resumeC: make(chan interface{}),
	}
}

func (c *coroutine) start(f coroutineFunc, pc *ProcessContext) interface{} {
	c.status = running
	go func() {
		defer func() {
			if r := recover(); r != nil {
				var ok bool
				c.panicE, ok = r.(error)
				if !ok {
					c.panicE = fmt.Errorf("coroutine panic: %v", r)
				}
			}

			c.status = stop
			c.yieldC <- nil
		}()

		f(pc)
	}()

	return <-c.yieldC
}

func (c *coroutine) yield(value interface{}) interface{} {
	c.status = suspend
	c.yieldC <- value
	return <-c.resumeC
}

func (c *coroutine) resume(value interface{}) interface{} {
	c.status = running
	c.resumeC <- value
	return <-c.yieldC
}

func (c *coroutine) getStatus() int32 {
	return c.status
}
