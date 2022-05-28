package breaker

import (
	"github.com/jeevic/lego/components/breakers/define"
	"math/rand"
	"sync"
	"time"
)

const (
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	letterIdxBits  = 6 // 6 bits to represent a letter index
	defaultRandLen = 8
	letterIdxMask  = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax   = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

type (
	EffectiveBreaker interface {
		// Name returns the name of the EffectiveBreaker.
		Name() string
		// Allow checks if the request is allowed.
		// If allowed, a promise will be returned, the caller needs to call promise.Accept()
		// on success, or call promise.Reject() on failure.
		// If not allow, ErrServiceUnavailable will be returned.
		Allow() (internalPromise, error)
		// Do runs the given request if the EffectiveBreaker accepts it.
		// Do returns an error instantly if the EffectiveBreaker rejects the request.
		// If a panic occurs in the request, the EffectiveBreaker handles it as an error
		// and causes the same panic again.
		Do(req ReqFunc) error
		// DoWithAcceptable runs the given request if the EffectiveBreaker accepts it.
		// DoWithAcceptable returns an error instantly if the EffectiveBreaker rejects the request.
		// If a panic occurs in the request, the EffectiveBreaker handles it as an error
		// and causes the same panic again.
		// acceptable checks if it's a successful call, even if the err is not nil.
		DoWithAcceptable(req ReqFunc, acceptable Acceptable) error
		// DoWithFallback runs the given request if the EffectiveBreaker accepts it.
		// DoWithFallback runs the fallback if the EffectiveBreaker rejects the request.
		// If a panic occurs in the request, the EffectiveBreaker handles it as an error
		// and causes the same panic again.
		DoWithFallback(req ReqFunc, fallback FallbackFunc) error
		// DoWithFallbackAcceptable runs the given request if the EffectiveBreaker accepts it.
		// DoWithFallbackAcceptable runs the fallback if the EffectiveBreaker rejects the request.
		// If a panic occurs in the request, the EffectiveBreaker handles it as an error
		// and causes the same panic again.
		// acceptable checks if it's a successful call, even if the err is not nil.
		DoWithFallbackAcceptable(req ReqFunc, fallback FallbackFunc, acceptable Acceptable) error
	}

	ReqFunc      func() error
	FallbackFunc func(err error) error
	Acceptable   func(err error) bool

	// Option defines the method to customize a EffectiveBreaker.
	Option func(breaker *circuitBreaker)

	internalPromise interface {
		Accept()
		Reject()
	}

	circuitBreaker struct {
		name        string
		breakerType define.BreakerType
		timeout     time.Duration
		internBreaker
	}

	internBreaker interface {
		allow() (internalPromise, error)
		doReq(req ReqFunc, fallback FallbackFunc, acceptable Acceptable, timeout time.Duration) error
	}

	lockedSource struct {
		source rand.Source
		lock   sync.Mutex
	}
)

var typeInitMap = map[define.BreakerType]internBreaker{
	define.DefaultBreaker: NewGoogleBreaker(),
	define.GoogleBreaker:  NewGoogleBreaker(),
}

func (ls *lockedSource) Int63() int64 {
	ls.lock.Lock()
	defer ls.lock.Unlock()
	return ls.source.Int63()
}

func newLockedSource(seed int64) *lockedSource {
	return &lockedSource{
		source: rand.NewSource(seed),
	}
}

var src = newLockedSource(time.Now().UnixNano())

func (cb *circuitBreaker) Name() string {
	return cb.name
}

func (cb *circuitBreaker) Allow() (internalPromise, error) {
	return cb.internBreaker.allow()
}

func (cb *circuitBreaker) Do(req ReqFunc) error {
	return cb.internBreaker.doReq(req, nil, defaultAcceptable, cb.timeout)
}

func (cb *circuitBreaker) DoWithAcceptable(req ReqFunc, acceptable Acceptable) error {
	return cb.internBreaker.doReq(req, nil, acceptable, cb.timeout)
}

func (cb *circuitBreaker) DoWithFallback(req ReqFunc, fallback FallbackFunc) error {
	return cb.internBreaker.doReq(req, fallback, defaultAcceptable, cb.timeout)
}

func (cb *circuitBreaker) DoWithFallbackAcceptable(req ReqFunc, fallback FallbackFunc,
	acceptable Acceptable) error {
	return cb.internBreaker.doReq(req, fallback, acceptable, cb.timeout)
}

func NewBreaker(opts ...Option) EffectiveBreaker {
	var b circuitBreaker
	for _, opt := range opts {
		opt(&b)
	}
	if len(b.name) == 0 {
		b.name = getRandName()
	}
	b.internBreaker = typeInitMap[b.breakerType]

	return &b
}

// WithName returns a function to set the name of a EffectiveBreaker.
func WithName(name string) Option {
	return func(b *circuitBreaker) {
		b.name = name
	}
}

// WithBreakerType 预留的，可能以后有多种不同类型的breaker returns a function to set the type of a EffectiveBreaker.
func WithBreakerType(breakerType define.BreakerType) Option {
	return func(b *circuitBreaker) {
		b.breakerType = breakerType
	}
}

// WithTimeout returns a function to set the timeout param of a EffectiveBreaker.
func WithTimeout(timeout time.Duration) Option {
	return func(b *circuitBreaker) {
		b.timeout = timeout
	}
}

func defaultAcceptable(err error) bool {
	return err == nil
}

func getRandName() string {
	b := make([]byte, defaultRandLen)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := defaultRandLen-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}
