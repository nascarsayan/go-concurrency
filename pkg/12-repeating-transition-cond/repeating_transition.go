// https://play.golang.org/p/HYiRtJcyaX9
package repeating_transition_cond

import (
	"fmt"
	"sync"
	"time"
)

type Idler struct {
	mu        sync.Mutex
	idle      sync.Cond
	busy      bool
	idleCount int64
}

func (i *Idler) AwaitIdle() {
	i.mu.Lock()
	defer i.mu.Unlock()
	idleCount := i.idleCount
	for i.busy && i.idleCount == idleCount {
		i.idle.Wait()
	}
	fmt.Printf("Waited for idle: %+v\n", i.idleCount)
}

func (i *Idler) SetBusy(b bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	wasBusy := i.busy
	i.busy = b
	if wasBusy && !i.busy {
		i.idleCount++
		fmt.Printf("Idle again for the %+vth time\n", i.idleCount)
		i.idle.Broadcast()
	}
}

func NewIdler() *Idler {
	i := &Idler{}
	i.idle.L = &i.mu
	return i
}

func RepeatingTransition() {
	i := NewIdler()
	i.SetBusy(true)
	go func() {
		time.Sleep(1 * time.Second)
		i.SetBusy(false)
	}()
	i.AwaitIdle()
	i.AwaitIdle()
}
