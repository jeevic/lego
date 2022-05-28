package breaker

import (
	"errors"
	"github.com/jeevic/lego/components/breakers/utils"
	"math"
	"time"
)

const (
	// 250ms for bucket duration
	window  = time.Second * 10
	buckets = 40

	//降低 K值 会使自适应限流算法更加激进（允许客户端在算法启动时拒绝更多本地请求）
	//增加 K值 会使自适应限流算法不再那么激进（允许服务端在算法启动时尝试接收更多的请求，与上面相反）
	k = 2
	// breaker失败，总次数 前几次 不会计入熔断器计数，延迟开启，只要失败n次之内不会触发熔断机制
	// 出错几率较小的场景建议这个给个初始值，越大熔断器就越不激进【如果小的话(失败/总次数)增加的快】
	protection = 2
)

// ErrServiceUnavailable is returned when the EffectiveBreaker state is open.
var ErrServiceUnavailable = errors.New("circuit breaker is open")

type (
	// googleBreaker is a netflixBreaker pattern from google.
	// see Client-Side Throttling section in https://landing.google.com/sre/sre-book/chapters/handling-overload/
	googleBreaker struct {
		k     float64
		stat  RollingWindow
		proba utils.Proba
	}
)

func NewGoogleBreaker() *googleBreaker {
	bucketDuration := time.Duration(int64(window) / int64(buckets))
	st := *NewRollingWindow(buckets, bucketDuration)
	return &googleBreaker{
		stat:  st,
		k:     k,
		proba: *utils.NewProba(),
	}
}

func (b *googleBreaker) accept() error {
	accepts, total := b.history()
	weightedAccepts := b.k * float64(accepts)
	// https://landing.google.com/sre/sre-book/chapters/handling-overload/#eq2101
	// 算法熔断概率： (requests-k*accepts)/(requests + 1)
	// 改进增加了一个protection，防止熔断器过快的启动并且熔断几率增加过快
	dropRatio := math.Max(0, (float64(total-protection)-weightedAccepts)/float64(total+1))
	//todo debug: print the probability of fusing
	//fmt.Printf("====接受次数：%v k*accept = %v，总次数：%v，熔断概率：%.2f====", accepts,b.k * float64(accepts), total, dropRatio)
	if dropRatio <= 0 {
		return nil
	}
	if b.proba.TrueOnProba(dropRatio) {
		return ErrServiceUnavailable
	}
	return nil
}

func (b *googleBreaker) allow() (internalPromise, error) {
	if err := b.accept(); err != nil {
		return nil, err
	}

	return googlePromise{
		b: b,
	}, nil
}

func (b *googleBreaker) doReq(req ReqFunc, fallback FallbackFunc, acceptable Acceptable, timeout time.Duration) error {
	if err := b.accept(); err != nil {
		if fallback != nil {
			return fallback(err)
		}
		return err
	}

	defer func() {
		if e := recover(); e != nil {
			b.markFailure()
			panic(e)
		}
	}()

	beforeExec := time.Now()
	if req == nil {
		return errors.New("request func can not equals nil")
	}
	err := req()
	// 没有出现timeout并且错误时可以接受的错误
	if !(timeout != 0 && time.Since(beforeExec) < timeout) && acceptable(err) {
		b.markSuccess()
	} else {
		b.markFailure()
	}
	return err
}

func (b *googleBreaker) markSuccess() {
	b.stat.Add(1)
}

func (b *googleBreaker) markFailure() {
	b.stat.Add(0)
}

func (b *googleBreaker) history() (accepts, total int64) {
	b.stat.Reduce(func(b *Bucket) {
		accepts += int64(b.Sum)
		total += b.Count
	})

	return
}

type googlePromise struct {
	b *googleBreaker
}

func (p googlePromise) Accept() {
	p.b.markSuccess()
}

func (p googlePromise) Reject() {
	p.b.markFailure()
}
