// https://go.dev/play/p/8m1i4IgeaIw
package wait_signal

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
)

type Item = int

type Queue struct {
	mu        sync.Mutex
	items     []Item
	itemAdded sync.Cond
}

func NewQueue() *Queue {
	q := new(Queue)
	q.itemAdded.L = &q.mu
	return q
}

func (q *Queue) Get() Item {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == 0 {
		q.itemAdded.Wait()
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *Queue) Put(item Item) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, item)
	// q.itemAdded.Signal()
	q.itemAdded.Broadcast()
}

func (q *Queue) GetMany(n int) []Item {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) < n {
		q.itemAdded.Wait()
	}
	items := q.items[:n:n]
	q.items = q.items[n:]
	return items
}

// randYield randomly yields the current goroutine, and makes way for some other.
// Not used much, as the Go scheduler usually handles goroutine scheduling efficiently.
func randYield() {
	if rand.Intn(3) == 0 {
		runtime.Gosched()
	}
}

func WaitSignal() {
	q := NewQueue()

	wg := sync.WaitGroup{}
	for n := 10; n > 0; n-- {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			randYield()
			items := q.GetMany(n)
			fmt.Printf("Retrieced %2d items : %+v\n", n, items)
		}(n)
	}

	for i := range 100 {
		q.Put(i)
		runtime.Gosched()
	}

	wg.Wait()
}
