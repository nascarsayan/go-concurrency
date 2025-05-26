package cond_respool

import (
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

type Pool struct {
	mu              sync.Mutex
	cond            sync.Cond
	numConns, limit int
	idle            []net.Conn
}

func NewPool(limit int) *Pool {
	p := new(Pool)
	p.limit = limit
	p.cond.L = &p.mu
	return p
}

func (p *Pool) Acquire() (net.Conn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for len(p.idle) == 0 && p.numConns >= p.limit {
		p.cond.Wait()
	}

	if len(p.idle) > 0 {
		// use cached conenction
		c := p.idle[len(p.idle)-1]
		p.idle = p.idle[:len(p.idle)-1]
		return c, nil
	}

	// create new connection
	c, err := dial()
	if err == nil {
		p.numConns++
	}
	return c, err
}

func (p *Pool) Release(c net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.idle = append(p.idle, c)
	p.cond.Signal()
}

func (p *Pool) Hijack(c net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.numConns--
	p.cond.Signal()
}

func randYield() {
	if rand.Intn(4) == 0 {
		runtime.Gosched()
	}
}

func ResourcePoolCond() {
	p := NewPool(3)

	wg := sync.WaitGroup{}
	for n := 10; n > 0; n-- {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			randYield()

			conn, err := p.Acquire()
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

			conn, err := p.Acquire()
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
