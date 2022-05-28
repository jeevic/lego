package breakers

import (
	"github.com/jeevic/lego/components/breakers/breaker"
	"github.com/jeevic/lego/components/breakers/define"
	"sync"
	"time"
)

var (
	lock     sync.RWMutex
	breakers = map[string]breaker.EffectiveBreaker{}
)

type Breaker struct {
	name          string
	timeout       time.Duration
	breakerType   define.BreakerType
	fallback      breaker.FallbackFunc
	acceptable    breaker.Acceptable
	activeBreaker breaker.EffectiveBreaker
}

func NewBreaker() *Breaker {
	return &Breaker{}
}

func (ob *Breaker) WithFallback(fallback breaker.FallbackFunc) *Breaker {
	ob.fallback = fallback
	return ob
}

func (ob *Breaker) WithAcceptable(acceptable breaker.Acceptable) *Breaker {
	ob.acceptable = acceptable
	return ob
}

func (ob *Breaker) WithTimeout(timeout time.Duration) *Breaker {
	ob.timeout = timeout
	return ob
}

func (ob *Breaker) WithBreakerType(bt define.BreakerType) *Breaker {
	ob.breakerType = bt
	return ob
}

func (ob *Breaker) GetOrBuild(name string) *Breaker {
	lock.RLock()
	bk, ok := breakers[name]
	if ok {
		ob.activeBreaker = bk
		lock.RUnlock()
		return ob
	}
	lock.RUnlock()

	lock.Lock()
	defer lock.Unlock()
	bk, ok = breakers[name]
	if !ok {
		bk = breaker.NewBreaker(breaker.WithName(name), breaker.WithTimeout(ob.timeout), breaker.WithBreakerType(ob.breakerType))
		breakers[name] = bk
		ob.activeBreaker = bk
	}
	return ob
}

func (ob *Breaker) Do(req breaker.ReqFunc) error {
	if ob.activeBreaker == nil {
		// 未初始化如果直接使用
		ob.activeBreaker = breaker.NewBreaker()
	}
	if ob.acceptable != nil && ob.fallback != nil {
		return ob.activeBreaker.DoWithFallbackAcceptable(req, ob.fallback, ob.acceptable)
	} else if ob.acceptable != nil {
		return ob.activeBreaker.DoWithAcceptable(req, ob.acceptable)
	} else if ob.fallback != nil {
		return ob.activeBreaker.DoWithFallback(req, ob.fallback)
	}
	return ob.activeBreaker.Do(req)
}
