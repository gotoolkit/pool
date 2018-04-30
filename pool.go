package pool

import (
	"errors"
	"io"
	"log"
	"sync"
)

type Factory func() (io.Closer, error)

// Pool 资源池
type Pool struct {
	m         sync.Mutex
	resources chan io.Closer
	factory   Factory
	closed    bool
}

var ErrPoolClosed = errors.New("Pool has been closed")

// New 创建资源池
func New(fn Factory, size uint) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("Size too small")
	}

	return &Pool{
		factory:   fn,
		resources: make(chan io.Closer, size),
	}, nil
}

// Acquire 获取资源
func (p *Pool) Acquire() (io.Closer, error) {
	select {
	case r, ok := <-p.resources:
		log.Println("Acquire:", "Shared Resource")
		if !ok {
			return nil, ErrPoolClosed
		}
		return r, nil
	}
	return nil, nil
}

// Release 释放资源
func (p *Pool) Release(r io.Closer) {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		r.Close()
		return
	}
	select {
	case p.resources <- r:
		log.Println("Release:", "In Queue")
	default:
		log.Println("Release:", "Closing")
		r.Close()
	}
}

// Close 关闭资源池
func (p *Pool) Close() {
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		return
	}
	p.closed = true
	close(p.resources)
	for r := range p.resources {
		r.Close()
	}
}
