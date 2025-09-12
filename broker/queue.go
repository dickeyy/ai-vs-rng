package broker

import (
	"container/list"
	"sync"
)

// TradeQueue is a thread-safe queue for trades.
type TradeQueue struct {
	mu    sync.Mutex
	queue *list.List
}

// NewTradeQueue creates and returns a new TradeQueue.
func NewTradeQueue() *TradeQueue {
	return &TradeQueue{
		queue: list.New(),
	}
}

// Enqueue adds a work item to the back of the queue.
func (q *TradeQueue) Enqueue(item *workItem) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue.PushBack(item)
}

// Dequeue removes and returns a work item from the front of the queue.
// It returns nil if the queue is empty.
func (q *TradeQueue) Dequeue() *workItem {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.queue.Len() == 0 {
		return nil
	}
	e := q.queue.Front()
	q.queue.Remove(e)
	return e.Value.(*workItem)
}

// IsEmpty returns true if the queue is empty, false otherwise.
func (q *TradeQueue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.queue.Len() == 0
}
