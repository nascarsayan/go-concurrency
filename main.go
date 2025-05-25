package main

import (
	_ "github.com/nascarsayan/go-concurrency/pkg/01-callback"
	_ "github.com/nascarsayan/go-concurrency/pkg/02-futures"
	_ "github.com/nascarsayan/go-concurrency/pkg/03-queue"
	_ "github.com/nascarsayan/go-concurrency/pkg/05-async-at-callsite"
	_ "github.com/nascarsayan/go-concurrency/pkg/06-async-internal"
	cond "github.com/nascarsayan/go-concurrency/pkg/07-cond"
)

func main() {
	// async.Callback()
	// futures.Futures()
	// queue.Queue()
	// async_at_callsite.AsyncAtCallsite(context.Background())
	// async_internal.AsyncInternal()
	cond.WaitSignal()
}
