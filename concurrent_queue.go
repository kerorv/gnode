package gnode

import "container/list"
import "sync"

type concurrentQueue struct {
	items *list.List
	mutex *sync.Mutex
}

func newConcurrentQueue() *concurrentQueue {
	return &concurrentQueue{
		items: list.New(),
		mutex: &sync.Mutex{},
	}
}

func (q *concurrentQueue) push(value interface{}) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.items.PushBack(value)
}

func (q *concurrentQueue) pop() interface{} {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	e := q.items.Front()
	if e != nil {
		q.items.Remove(e)
		return e.Value
	}

	return nil
}

func (q *concurrentQueue) count() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.items.Len()
}
