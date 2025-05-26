// https://play.golang.org/p/rzSXpophC_p
package queue_chan_state

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
)

type Item = int

type waiter struct {
	n int
	c chan []Item
}

type state struct {
	items []Item
	wait  []waiter
}

type Queue struct {
	s chan state
}

func NewQueue() *Queue {
	s := make(chan state, 1)
	s <- state{}
	return &Queue{s}
}

func (q *Queue) GetMany(n int) []Item {
	s := <-q.s
	if len(s.items) >= n && len(s.wait) == 0 {
		items := s.items[:n:n]
		s.items = s.items[n:]
		q.s <- s
		return items
	}
	c := make(chan []Item)
	s.wait = append(s.wait, waiter{n, c})
	q.s <- s
	return <-c
}

func (q *Queue) Put(item Item) {
	s := <-q.s
	s.items = append(s.items, item)
	for len(s.wait) > 0 {
		w := s.wait[0]
		if w.n > len(s.items) {
			break
		}
		w.c <- s.items[:w.n:w.n]
		s.items = s.items[w.n:]
		s.wait = s.wait[1:]
	}
	q.s <- s
}

func randYield() {
	if rand.Intn(3) == 0 {
		runtime.Gosched()
	}
}

func QueueChanWithState() {
	// TODO: Add cancellation with context.
	q := NewQueue()

	var wg sync.WaitGroup
	for n := 10; n > 0; n-- {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			randYield()
			items := q.GetMany(n)
			fmt.Printf("%2d: %2d\n", n, items)
		}(n)
	}

	for i := range 100 {
		q.Put(i)
		runtime.Gosched()
	}

	wg.Wait()
}
