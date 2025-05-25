// https://play.golang.org/p/jlrZshjnJ9z
package async

import (
	"fmt"
	"os"
	"sync/atomic"
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

// send callback function here.
// Fetch creates a new thread for doing some slow operation, and runs the callback `f` on the output.
// It's an asynchronous function and returns immediately.
func Fetch(name string, f func(Item)) {
	go func() {
		item := items[name]
		doSlowThing()
		f(item)
	}()
}

func Callback() {
	start := time.Now()

	n := int32(0)

	Fetch("a", func(i Item) {
		fmt.Printf("Fetched item: %+v\n", i)
		if atomic.AddInt32(&n, 1) == 2 {
			fmt.Printf("All items have been processed, total time = %+v\n", time.Since(start))
			os.Exit(0)
		}
	})

	Fetch("b", func(i Item) {
		fmt.Printf("Fetched item: %+v\n", i)
		if atomic.AddInt32(&n, 1) == 2 {
			fmt.Printf("All items have been processed, total time = %+v\n", time.Since(start))
			os.Exit(0)
		}
	})

	fmt.Println("Waiting for items to be processed...")
	// Wait for all goroutines to finish
	select {}
}
