package main

import (
	_ "github.com/nascarsayan/go-concurrency/pkg/01-callback"
	_ "github.com/nascarsayan/go-concurrency/pkg/02-futures"
	_ "github.com/nascarsayan/go-concurrency/pkg/03-queue"
	_ "github.com/nascarsayan/go-concurrency/pkg/05-async-at-callsite"
	_ "github.com/nascarsayan/go-concurrency/pkg/06-async-internal"
	_ "github.com/nascarsayan/go-concurrency/pkg/07-queue-cond"
	_ "github.com/nascarsayan/go-concurrency/pkg/08-respool-cond"
	_ "github.com/nascarsayan/go-concurrency/pkg/10-queue-chan"
	chan_queue_state "github.com/nascarsayan/go-concurrency/pkg/11-queue-chan-state"
)

func main() {
	// async.Callback()
	// futures.Futures()
	// queue.Queue()
	// async_at_callsite.AsyncAtCallsite(context.Background())
	// async_internal.AsyncInternal()
	// cond.WaitSignal()
	// cond_respool.ResourcePoolCond()
	// chan_respool.ResourcePoolChan()
	// chan_queue.QueueChan()
	chan_queue_state.QueueChanWithState()
}
