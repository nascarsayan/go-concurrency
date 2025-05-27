// https://go.dev/play/p/4WG59Juxjch
package repeating_transition_chan

import (
	"context"
	"fmt"
	"time"
)

type Idler struct {
	// next holds a channel if busy, nil if idle.
	// The channel is closed when busy-to-idle transition occurs.
	next chan chan struct{}
}

func (i *Idler) AwaitIdle(ctx context.Context) error {
	idle := <-i.next
	i.next <- idle
	if idle != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-idle:
		}
	}
	return nil
}

func (i *Idler) SetBusy(b bool) {
	idle := <-i.next
	if b && idle == nil {
		idle = make(chan struct{})
	} else if !b && idle != nil {
		close(idle)
		idle = nil
	}
	i.next <- idle
}

func NewIdler() *Idler {
	next := make(chan chan struct{}, 1)
	next <- nil
	return &Idler{next}
}

func RepeatingTransitionChan() {
	i := NewIdler()
	i.SetBusy(true)
	go func() {
		time.Sleep(1 * time.Second)
		i.SetBusy(false)
	}()

	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	fmt.Println(i.AwaitIdle(ctx))
	fmt.Println(time.Since(start))

	fmt.Println(i.AwaitIdle(context.Background()))
	fmt.Println(time.Since(start))
}
