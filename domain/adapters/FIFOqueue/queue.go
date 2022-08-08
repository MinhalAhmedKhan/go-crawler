package FIFOqueue

import (
	"sync/atomic"

	"github.com/antigloss/go/concurrent/container/queue"
)

type FIFOQueue struct {
	queue *queue.LockfreeQueue
	size  uint64
}

func New() *FIFOQueue {
	return &FIFOQueue{
		queue: queue.NewLockfreeQueue(),
		size:  0,
	}
}

func (q *FIFOQueue) Push(v interface{}) error {
	atomic.AddUint64(&q.size, 1)
	q.queue.Push(v)
	return nil
}

func (q *FIFOQueue) Pop() (interface{}, error) {
	atomic.AddUint64(&q.size, ^uint64(0))
	return q.queue.Pop(), nil
}
