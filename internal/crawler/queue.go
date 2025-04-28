package crawler

import (
	"container/list"
	"sync"
)

// URLQueue implements a thread-safe FIFO queue for URLs to be crawled
type URLQueue struct {
	queue  *list.List
	mu     sync.Mutex
	cond   *sync.Cond
	closed bool
}

// NewURLQueue creates a new URLQueue instance
func NewURLQueue() *URLQueue {
	q := &URLQueue{
		queue: list.New(),
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Enqueue adds a URL to the end of the queue
func (q *URLQueue) Enqueue(url string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return
	}

	q.queue.PushBack(url)
	q.cond.Signal() // Notify waiting goroutines that a new item is available
}

// Dequeue removes and returns a URL from the front of the queue
// Blocks if queue is empty until an item is available or queue is closed
func (q *URLQueue) Dequeue() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Wait until there's an item or the queue is closed
	for q.queue.Len() == 0 && !q.closed {
		q.cond.Wait()
	}

	// If queue is closed and empty, return false
	if q.closed && q.queue.Len() == 0 {
		return "", false
	}

	// Get the front element
	element := q.queue.Front()
	url := element.Value.(string)
	q.queue.Remove(element)

	return url, true
}

// Close the queue and wake up all waiting goroutines
func (q *URLQueue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.closed = true
	q.cond.Broadcast() // Wake up all waiting goroutines
}

// Size returns the current number of items in the queue
func (q *URLQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.queue.Len()
}

// IsEmpty returns true if the queue is empty
func (q *URLQueue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.queue.Len() == 0
}

// IsClosed returns true if the queue has been closed
func (q *URLQueue) IsClosed() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.closed
}