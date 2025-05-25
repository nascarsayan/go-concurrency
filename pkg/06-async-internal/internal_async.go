// https://go.dev/play/p/zbP4cko_rTI
package async_internal

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
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

func Match(ctx context.Context, pattern string) (names []string, err error) {
	for name := range items {
		if ok, _ := filepath.Match(pattern, name); ok {
			names = append(names, name)
		}
	}
	return names, nil
}

func GlobQuery(ctx context.Context, pattern string) ([]Item, error) {
	names, err := Match(ctx, pattern)
	if err != nil {
		return nil, err
	}

	ch := make(chan Item)
	wg := sync.WaitGroup{}

	for _, name := range names {
		wg.Add(1)
		go func(name string) (err error) {
			defer wg.Done()
			item, err := Fetch(ctx, name)
			if err == nil {
				ch <- item
			}
			return err
		}(name)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var items []Item
	for item := range ch {
		items = append(items, item)
	}
	return items, nil
}

func AsyncInternal() {
	start := time.Now()
	fmt.Println(GlobQuery(context.Background(), "[ab]*"))
	fmt.Println(time.Since(start))
}
