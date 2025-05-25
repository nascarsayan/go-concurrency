// https://go.dev/play/p/zbP4cko_rTI
package async_at_callsite

import (
	"context"
	"errors"
	"fmt"
	"sync"
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
	time.Sleep(10 * time.Millisecond)
}

// Fetch is synchronous, and it returns the requested item.
func Fetch(ctx context.Context, name string) (Item, error) {
	doSlowThing()
	item, ok := items[name]
	if !ok {
		return Item{}, errors.New("item not found")
	}
	return item, nil
}

func consume(items []Item) {
	fmt.Printf("%+v", items)
}

func AsyncAtCallsite(ctx context.Context) {
	toFetch := []string{"a", "b"}
	items := make([]Item, len(toFetch))
	wg := sync.WaitGroup{}
	for idx, name := range toFetch {
		idx, name := idx, name
		go func() (err error) {
			defer wg.Done()
			val, err := Fetch(ctx, name)
			items[idx] = val
			return err
		}()
		wg.Add(1)
	}
	wg.Wait()
	consume(items)
}
