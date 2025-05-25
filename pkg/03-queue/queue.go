// https://play.golang.org/p/GpnC3KgwlT0
package queue

import (
	"fmt"
	"path/filepath"
)

type Item struct {
	kind string
}

var items = map[string]Item{
	"a": {"alligator"},
	"b": {"bear"},
}

// Glob finds all the items with names matching the pattern
// and sends them on the returned channel.
// It closes the channel when all items have been sent.
func Glob(pattern string) <-chan Item {
	ch := make(chan Item)
	go func() {
		defer close(ch)
		for name, item := range items {
			if mat, _ := filepath.Match(pattern, name); !mat {
				continue
			}
			ch <- item
		}
	}()
	return ch
}

func Queue() {
	for item := range Glob("[ab]*") {
		fmt.Printf("%+v\n", item)
	}
}
