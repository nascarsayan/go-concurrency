// https://play.golang.org/p/uvx8vFSQ2f0
package queue_chan

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
)

type Item = int

type Queue struct {
	items chan []Item // Non-empty list of items
	empty chan bool   // Queue is empty
}

func NewQueue() *Queue {
	q := new(Queue)
	q.empty = make(chan bool, 1)
	q.empty <- true
	q.items = make(chan []Item, 1)
	return q
}

func (q *Queue) Get() Item {
	items := <-q.items
	item := items[0]
	items = items[1:]
	if len(items) == 0 {
		q.empty <- true
	} else {
		q.items <- items
	}
	return item
}

func (q *Queue) GetWithCtx(ctx context.Context) (Item, error) {
	var items []Item
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case items = <-q.items:
	}

	item := items[0]
	items = items[1:]
	if len(items) == 0 {
		q.empty <- true
	} else {
		q.items <- items
	}
	return item, nil
}

func (q *Queue) Put(item Item) {
	var items []Item
	select {
	case items = <-q.items:
	case <-q.empty:
	}
	items = append(items, item)
	q.items <- items
}

func randYield() {
	if rand.Intn(2) == 0 {
		runtime.Gosched()
	}
}

func randCancel() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	if rand.Intn(4) == 0 {
		cancel()
	} else {
		_ = cancel
	}
	return ctx
}

func QueueChan() {
	q := NewQueue()

	var wg sync.WaitGroup
	for n := 20; n > 0; n-- {
		wg.Add(1)
		go func(n int) {
			randYield()
			item := q.Get()
			fmt.Printf("Thread %2d: item %2d\n", n, item)
			wg.Done()
		}(n)
	}

	for i := range 100 {
		q.Put(i)
		randYield()
	}

	wg.Wait()
}
