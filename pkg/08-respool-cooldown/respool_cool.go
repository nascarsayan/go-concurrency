// https://go.dev/play/p/j_OmiKuyWo8
package repool_cooldown

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
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
		fmt.Printf("[dial] Listening on %s\n", addr.String())
		go func() {
			for {
				in, err := ln.Accept()
				if err != nil {
					fmt.Printf("[dial] Accept error: %v\n", err)
					return
				}
				fmt.Printf("[dial] Accepted connection from %s\n", in.RemoteAddr())
				go io.Copy(os.Stdout, in)
			}
		}()
	})
	fmt.Printf("[dial] Dialing %s\n", addr.String())
	return net.Dial(addr.Network(), addr.String())
}

type Pool struct {
	mu             sync.Mutex
	cool           sync.Cond
	idle           []net.Conn
	limit, numConn int
	hitLimit       bool
}

func NewPool(limit int) *Pool {
	p := new(Pool)
	p.cool.L = &p.mu
	p.limit = limit
	return p
}

func (p *Pool) Release(c net.Conn) {
	fmt.Printf("[Release] Releasing connection %v\n", c.RemoteAddr())
	p.mu.Lock()
	defer p.mu.Unlock()
	p.idle = append(p.idle, c)
	p.numConn--
	fmt.Printf("[Release] numConn: %d, idle: %d\n", p.numConn, len(p.idle))
	if p.hitLimit && p.numConn == 0 {
		fmt.Printf("[Release] hitLimit reset, broadcasting\n")
		p.hitLimit = false
		p.cool.Broadcast()
	}
}

func (p *Pool) Acquire() (net.Conn, error) {
	fmt.Printf("[Acquire] Attempting to acquire connection\n")
	p.mu.Lock()
	defer p.mu.Unlock()
	defer func() {
		if p.numConn == p.limit {
			fmt.Printf("[Acquire] hitLimit reached (%d)\n", p.limit)
			p.hitLimit = true
		}
	}()
	for p.hitLimit {
		fmt.Printf("[Acquire] hitLimit true, waiting\n")
		p.cool.Wait()
	}
	if len(p.idle) == 0 {
		fmt.Printf("[Acquire] No idle connections, dialing new\n")
		c, err := dial()
		if err == nil {
			p.numConn++
			fmt.Printf("[Acquire] New connection acquired, numConn: %d\n", p.numConn)
		} else {
			fmt.Printf("[Acquire] Dial error: %v\n", err)
		}
		return c, err
	}
	p.numConn++
	c := p.idle[0]
	p.idle = p.idle[1:]
	fmt.Printf("[Acquire] Reusing idle connection, numConn: %d, idle left: %d\n", p.numConn, len(p.idle))
	return c, nil
}

func ResourcePoolCool() {
	n := 5
	p := NewPool(n)
	wg := sync.WaitGroup{}

	conns := make(chan net.Conn)
	for i := range n * 3 {
		wg.Add(1)
		fmt.Printf("[main] Acquire %+v\n", i)
		go func(i int) {
			defer wg.Done()
			c, _ := p.Acquire()
			conns <- c
		}(i)
	}

	go func() {
		wg.Wait()
		close(conns)
	}()

	wg2 := sync.WaitGroup{}
	for c := range conns {
		wg2.Add(1)
		go func(c net.Conn) {
			defer wg2.Done()
			p.Release(c)
		}(c)
		time.Sleep(10 * time.Millisecond)
	}

	wg2.Wait()

}
