package broker

import (
	"container/list"
	"sync"

	"github.com/dickeyy/cis-320/types"
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

// Enqueue adds a trade to the back of the queue.
func (q *TradeQueue) Enqueue(trade *types.Trade) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue.PushBack(trade)
}

// Dequeue removes and returns a trade from the front of the queue.
// It returns nil if the queue is empty.
func (q *TradeQueue) Dequeue() *types.Trade {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.queue.Len() == 0 {
		return nil
	}
	e := q.queue.Front()
	q.queue.Remove(e)
	return e.Value.(*types.Trade)
}

// IsEmpty returns true if the queue is empty, false otherwise.
func (q *TradeQueue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.queue.Len() == 0
}
