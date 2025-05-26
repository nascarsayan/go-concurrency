// https://go.dev/play/p/j_OmiKuyWo8
package repool_chan

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"runtime"
	"sync"
)

var (
	addr     net.Addr
	addrOnce sync.Once
)

func dial() (net.Conn, error) {
	addrOnce.Do(func() {
		ln, err := net.Listen("tcp", ":0")
		if err != nil {
			panic(err)
		}
		addr = ln.Addr()
		go func() {
			for {
				in, err := ln.Accept()
				if err != nil {
					return
				}
				go io.Copy(os.Stdout, in)
			}
		}()
	})
	return net.Dial(addr.Network(), addr.String())
}

type token struct{}

type Pool struct {
	sem  chan token
	idle chan net.Conn
}

func NewPool(limit int) *Pool {
	sem := make(chan token, limit)
	idle := make(chan net.Conn, limit)
	return &Pool{sem: sem, idle: idle}
}

func (p *Pool) Release(c net.Conn) {
	p.idle <- c
}

func (p *Pool) Hijack(c net.Conn) {
	<-p.sem
}

func (p *Pool) Acquire(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case c := <-p.idle:
		return c, nil
	case p.sem <- token{}:
		conn, err := dial()
		if err != nil {
			<-p.sem
		}
		return conn, err
	}
}

func randYield() {
	if rand.Intn(4) == 0 {
		runtime.Gosched()
	}
}

func ResourcePoolChan() {
	p := NewPool(3)

	wg := sync.WaitGroup{}
	for n := 10; n > 0; n-- {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			randYield()

			conn, err := p.Acquire(context.Background())
			if err != nil {
				panic(err)
			}
			defer p.Release(conn)

			fmt.Fprintf(conn, "Hello from goroutine %d on connection %p!\n", n, conn)
			runtime.Gosched()
		}(n)
	}

	for n := 4; n > 0; n-- {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			randYield()

			conn, err := p.Acquire(context.Background())
			if err != nil {
				panic(err)
			}
			defer p.Hijack(conn)

			fmt.Fprintf(conn, "Goodbye from hijacked connection %p!\n", conn)
			runtime.Gosched()
		}(n)
	}

	wg.Wait()
}
