package godis

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis"
	"github.com/go-zookeeper/zk"
	"github.com/pkg/errors"

	"github.com/jeevic/lego/pkg/app"
)

// PooledObject is a pack struct for addr and the appropriate redis pool.
type PooledObject struct {
	Addr   string
	Client *redis.Client
}

// NewPooledObject return the pack struct for addr and the appropriate redis pool.
func NewPooledObject(addr string, client *redis.Client) *PooledObject {
	pooledObject := &PooledObject{
		Addr:   addr,
		Client: client,
	}
	return pooledObject
}

// ProxyInfo is represent the redis proxy instance is online or not.
type ProxyInfo struct {
	Addr  string `json:"addr"`
	State string `json:"state"`
}

// RoundRobinPool is a round-robin redis client pool for connecting multiple codis proxies based on
// zookeeper-go and redis-go.
type RoundRobinPool struct {
	zkConn             *zk.Conn
	zkAddr             []string
	zkSessionTimeoutMs time.Duration
	zkProxyDir         string
	pools              atomic.Value
	childCh            <-chan zk.Event
	childrenData       atomic.Value
	options            redis.Options
	nextIdx            int64
	stopChan           chan struct{}
	rwMutex            sync.RWMutex
}

// NewRoundRobinPool return a round-robin redis client pool specified by
// zk client and redis options.
func NewRoundRobinPool(zkConn *zk.Conn, zkProxyDir string, options redis.Options) (*RoundRobinPool, error) {
	pool := &RoundRobinPool{
		zkConn:     zkConn,
		zkProxyDir: zkProxyDir,
		nextIdx:    -1,
		pools:      atomic.Value{},
		options:    options,
		stopChan:   make(chan struct{}),
		rwMutex:    sync.RWMutex{},
	}
	pool.pools.Store([]*PooledObject{})
	_, _, childCh, err := zkConn.ChildrenW(zkProxyDir)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to watch %s", zkProxyDir))
	}
	pool.childCh = childCh
	pool.resetPools()

	go pool.watch()
	return pool, nil
}

//设置pools
func (p *RoundRobinPool) resetPools() {
	children, _, err := p.zkConn.Children(p.zkProxyDir)
	if err != nil {
		return
	}
	childrenData := make([]string, 0)
	for _, child := range children {
		data, _, err := p.zkConn.Get(p.zkProxyDir + "/" + child)
		if err != nil {
			continue
		}
		childrenData = append(childrenData, (string)(data))
	}
	sort.Strings(childrenData)

	pools := p.pools.Load().([]*PooledObject)
	addr2Pool := make(map[string]*PooledObject, len(pools))
	for _, pool := range pools {
		addr2Pool[pool.Addr] = pool
	}
	newPools := make([]*PooledObject, 0)
	for _, childData := range childrenData {
		proxyInfo := ProxyInfo{}
		err := json.Unmarshal([]byte(childData), &proxyInfo)
		if err != nil {
			continue
		}
		if proxyInfo.State != "online" {
			continue
		}
		addr := proxyInfo.Addr
		if pooledObject, ok := addr2Pool[addr]; ok {
			newPools = append(newPools, pooledObject)
			delete(addr2Pool, addr)
		} else {
			options := p.cloneOptions()
			options.Addr = addr
			options.Network = "tcp"
			pooledObject := NewPooledObject(
				addr,
				redis.NewClient(&options),
			)
			newPools = append(newPools, pooledObject)
		}
	}
	p.rwMutex.Lock()
	defer p.rwMutex.Unlock()
	p.pools.Store(newPools)
	for _, pooledObject := range addr2Pool {
		_ = pooledObject.Client.Close()
	}
}

// GetClient can get a redis client from pool with round-robin policy.
// It's safe for concurrent use by multiple goroutines.
func (p *RoundRobinPool) GetClient() (*redis.Client, error) {
	p.rwMutex.RLock()
	defer p.rwMutex.RUnlock()
	pools := p.pools.Load().([]*PooledObject)
	if len(pools) <= 0 {
		return nil, errors.New("proxy list  empty")
	}
	for {
		current := atomic.LoadInt64(&p.nextIdx)
		var next int64
		if (current) >= (int64)(len(pools))-1 {
			next = 0
		} else {
			next = current + 1
		}
		if atomic.CompareAndSwapInt64(&p.nextIdx, current, next) {
			return pools[next].Client, nil
		}
	}

}

func (p *RoundRobinPool) watch() {
	for {
		select {
		case event := <-p.childCh:
			app.App.GetLogger().Infof("godis  zookeeper change happened event:%+v", event)
			if event.State == zk.StateDisconnected {
				_, _, p.childCh, _ = p.zkConn.ChildrenW(p.zkProxyDir)
				continue
			}
			if event.Path != p.zkProxyDir {
				continue
			}
			if event.Type == zk.EventNodeChildrenChanged {
				p.resetPools()
				_, _, p.childCh, _ = p.zkConn.ChildrenW(p.zkProxyDir)
			}
		case <-p.stopChan:
			break
		}
	}
}

func (p *RoundRobinPool) cloneOptions() redis.Options {
	options := redis.Options{
		Network:            p.options.Network,
		Addr:               p.options.Addr,
		Dialer:             p.options.Dialer,
		OnConnect:          p.options.OnConnect,
		Password:           p.options.Password,
		DB:                 p.options.DB,
		MaxRetries:         p.options.MaxRetries,
		MinRetryBackoff:    p.options.MinRetryBackoff,
		MaxRetryBackoff:    p.options.MaxRetryBackoff,
		DialTimeout:        p.options.DialTimeout,
		ReadTimeout:        p.options.ReadTimeout,
		WriteTimeout:       p.options.WriteTimeout,
		PoolSize:           p.options.PoolSize,
		PoolTimeout:        p.options.PoolTimeout,
		IdleTimeout:        p.options.IdleTimeout,
		IdleCheckFrequency: p.options.IdleCheckFrequency,
		TLSConfig:          p.options.TLSConfig,
	}
	return options
}

// Close closes the pool, releasing all resources except zookeeper client.
func (p *RoundRobinPool) Close() {
	pools := p.pools.Load().([]*PooledObject)
	for _, pool := range pools {
		_ = pool.Client.Close()
	}
	//关闭watcher
	p.stopChan <- struct{}{}
	//关闭zk
	if p.zkConn != nil {
		p.zkConn.Close()
	}
}
