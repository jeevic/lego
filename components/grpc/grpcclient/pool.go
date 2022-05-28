package grpcclient

import (
	"sync"
	"sync/atomic"
)

type Pool struct {
	option *Options
	//address
	target string
	//容量
	capacity int64
	//
	next int64
	sync.Mutex
	clients []*GrpcClient
}

func NewPool(target string, options ...Option) (*Pool, error) {
	opts := NewOptions(options...)

	p := &Pool{
		option:   opts,
		target:   target,
		capacity: opts.PoolCap,
		clients:  make([]*GrpcClient, opts.PoolCap),
	}

	err := p.Init()
	if err != nil {
		return nil, err
	}
	return p, nil
}

//init connect
func (p *Pool) Init() error {
	for idx, _ := range p.clients {
		c, err := p.connect()
		if err != nil {
			return err
		}
		p.clients[idx] = c
	}
	return nil
}

func (p *Pool) connect() (*GrpcClient, error) {
	c, err := NewClient(p.target, p.option)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (p *Pool) GetClient() (*GrpcClient, error) {
	var (
		idx int64
		nxt int64
		err error
	)

	nxt = atomic.AddInt64(&p.next, 1)
	idx = nxt % p.capacity

	c := p.clients[idx]

	if c != nil && c.CheckState() == nil {
		return c, nil
	}

	//connect is close
	if c != nil {
		c.Close()
	}

	p.Lock()
	defer p.Unlock()

	// double check
	c = p.clients[idx]
	if c != nil && c.CheckState() == nil {
		return c, nil
	}

	client, err := p.connect()
	if err != nil {
		return nil, err
	}
	p.clients[idx] = client
	return client, nil
}

// close all
func (p *Pool) Close() {
	p.Lock()
	defer p.Unlock()

	for _, c := range p.clients {
		if c == nil || c.GetConn() == nil {
			continue
		}
		c.Close()
	}
}
