// https://play.golang.org/p/v_IGf8tU3UT
package futures

import (
	"fmt"
	"time"
)

type Item struct {
	kind string
}

var items = map[string]Item{
	"a": {"alligator"},
	"b": {"bear"},
}

func doSlowThing() {
	time.Sleep(1000 * time.Millisecond)
}

// Fetch immediately returns a channel. It performs the slow operation and
// eventually fetches the requested item and sends via the channel.
func Fetch(name string) <-chan Item {
	ch := make(chan Item)
	go func() {
		doSlowThing()
		item, ok := items[name]
		if !ok {
			close(ch)
			return
		}
		ch <- item
	}()
	return ch
}

func consume(a, b Item) {
	fmt.Printf("%+v %+v\n", a, b)
}

func Futures() {
	start := time.Now()
	a := Fetch("a")
	b := Fetch("b")
	consume(<-a, <-b)
	fmt.Printf("All items have been processed, total time = %+v\n", time.Since(start))
}
