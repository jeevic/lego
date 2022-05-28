package breakers

import (
	"errors"
	"fmt"
	"github.com/jeevic/lego/components/breakers/breaker"
	"github.com/jeevic/lego/components/breakers/define"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

var divisionZeroErr = errors.New("dividend can not equals 0")

func division(a int, b int) (res int, err error) {
	time.Sleep(20 * time.Millisecond)
	if b == 0 {
		return 0, divisionZeroErr
	}
	return a / b, nil
}

func divisionPanic(a int, b int) (res int, err error) {
	return a / b, nil
}

func TestNormalBreakerData(t *testing.T) {
	normalReqFunc := func() error {
		res, err := division(1, 2)
		time.Sleep(time.Millisecond * 20)
		if err != nil {
			return err
		}
		_ = res
		return nil
	}

	tests := []struct {
		name     string
		req      func() error
		accept   breaker.Acceptable
		fallback breaker.FallbackFunc
		timeout  time.Duration
	}{
		{name: "test accept",
			req: normalReqFunc,
			accept: func(err error) bool {
				if errors.Is(err, divisionZeroErr) {
					return false
				}
				return true
			}},
		{name: "test fallback", req: normalReqFunc,
			fallback: func(err error) error {
				fmt.Printf("breaker has been triggered")
				return nil
			}},
		{name: "test timeout", req: normalReqFunc,
			accept: func(err error) bool {
				return true
			}, fallback: func(err error) error {
				return nil
			}, timeout: time.Millisecond * 50},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			optBreaker := NewBreaker().WithFallback(test.fallback).
				WithAcceptable(test.accept).WithTimeout(test.timeout).GetOrBuild(test.name)
			for i := 0; i < 500; i++ {
				err := optBreaker.Do(test.req)
				if err != nil {
					fmt.Printf("%v\n", err)
				}
			}
		})
	}
}

func TestAbnormalBreakerData(t *testing.T) {

	normalReqFunc := func() error {
		res, err := division(1, 2)
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 30)
		fmt.Printf(strconv.Itoa(res) + "sleep::sleep")
		rand.Int31n(20)
		return nil
	}

	abnormalReqFunc := func() error {
		res, err := division(1, 0)
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 5)
		fmt.Printf(strconv.Itoa(res) + "sleep::sleep")
		return nil
	}

	tests := []struct {
		name     string
		req      func() error
		accept   breaker.Acceptable
		fallback breaker.FallbackFunc
		timeout  time.Duration
	}{
		{name: "all nil"},
		{name: "test accept",
			req: abnormalReqFunc,
			accept: func(err error) bool {
				// 方法返回true表明没错或者错误可允许，不计入熔断器计数
				if errors.Is(err, divisionZeroErr) {
					return false
				}
				return true
			}},
		{name: "test fallback", req: abnormalReqFunc,
			fallback: func(err error) error {
				fmt.Printf("breaker has been triggered, so reject execute fallback")
				return nil
			}},
		{name: "test timeout", req: normalReqFunc,
			accept: func(err error) bool {
				// 方法返回true表明没错或者错误可允许，不计入熔断器计数
				if errors.Is(err, divisionZeroErr) {
					return true
				}
				return false
			}, fallback: func(err error) error {
				fmt.Printf("normal but timeout so breaker triggerd, so reject to execute fallback")
				return nil
			}, timeout: time.Millisecond * 10}, // 10毫秒以上算超时，请求继续但是熔断器计数打开
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			optBreaker := NewBreaker().WithFallback(test.fallback).
				WithAcceptable(test.accept).WithTimeout(test.timeout).GetOrBuild(test.name)
			for i := 0; i < 200; i++ {
				fmt.Printf("第%vrange：：", i)
				err := optBreaker.Do(test.req)
				if err != nil {
					fmt.Printf("%v\n", err)
				}
				fmt.Println()
			}
		})
	}
}

func TestRecoverAbnormalBreakerData(t *testing.T) {

	normalReqFunc := func() error {
		res, err := division(1, 2)
		if err != nil {
			return err
		}
		fmt.Printf(strconv.Itoa(res) + "sleep::sleep")
		return nil
	}

	abnormalReqFunc := func() error {
		res, err := division(1, 0)
		if err != nil {
			return err
		}
		fmt.Printf(strconv.Itoa(res) + "sleep::sleep /0")
		return nil
	}

	tests := []struct {
		name     string
		req      []func() error
		accept   breaker.Acceptable
		fallback breaker.FallbackFunc
		timeout  time.Duration
	}{
		{name: "all nil"},
		{name: "test accept",
			req: []func() error{normalReqFunc, abnormalReqFunc},
			accept: func(err error) bool {
				// 方法返回true表明没错或者错误可允许，不计入熔断器计数
				if errors.Is(err, divisionZeroErr) {
					return false
				}
				return true
			}},
		{name: "test fallback", req: []func() error{normalReqFunc, abnormalReqFunc},
			fallback: func(err error) error {
				fmt.Printf("breaker has been triggered, so reject execute fallback")
				return nil
			}},
		{name: "test timeout", req: []func() error{normalReqFunc, abnormalReqFunc},
			accept: func(err error) bool {
				// 方法返回true表明没错或者错误可允许，不计入熔断器计数
				if errors.Is(err, divisionZeroErr) {
					return true
				}
				return false
			}, fallback: func(err error) error {
				fmt.Printf("normal but timeout so breaker triggerd, so reject to execute fallback")
				return nil
			}, timeout: time.Millisecond * 20}, // 10毫秒以上算超时，请求继续但是熔断器计数打开
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			optBreaker := NewBreaker().WithFallback(test.fallback).
				WithAcceptable(test.accept).WithTimeout(test.timeout).GetOrBuild(test.name)
			for i := 0; i < 500; i++ {
				n := rand.Int31n(500)
				fmt.Printf("是否出错 %v 第%vrange：：", n > int32(i), i)
				var err error
				if n < int32(i) {
					err = optBreaker.Do(test.req[0])
				} else {
					err = optBreaker.Do(test.req[1])
				}
				if err != nil {
					fmt.Printf("%v\n", err)
				}
				fmt.Println()
			}
		})
	}
}

func TestPanicNormalBreakerData(t *testing.T) {
	panicReqFunc := func() error {
		res, err := divisionPanic(1, 0)
		time.Sleep(time.Millisecond * 20)
		if err != nil {
			return err
		}
		_ = res
		return nil
	}

	tests := []struct {
		name     string
		req      func() error
		accept   breaker.Acceptable
		fallback breaker.FallbackFunc
		timeout  time.Duration
	}{
		{name: "all nil"},
		{name: "test accept",
			req: panicReqFunc,
			accept: func(err error) bool {
				if errors.Is(err, divisionZeroErr) {
					return false
				}
				return true
			}},
		{name: "test fallback", req: panicReqFunc,
			fallback: func(err error) error {
				fmt.Printf("breaker has been triggered")
				return nil
			}},
		{name: "test timeout", req: panicReqFunc,
			accept: func(err error) bool {
				return true
			}, fallback: func(err error) error {
				return nil
			}, timeout: 10 * time.Millisecond},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			optBreaker := NewBreaker().WithFallback(test.fallback).
				WithAcceptable(test.accept).WithTimeout(test.timeout).GetOrBuild(test.name)
			for i := 0; i < 500; i++ {
				func() {
					// 方法返回前执行
					defer func() {
						if e := recover(); e != nil {
							fmt.Printf("发生panic此处recover")
						}
					}()
					err := optBreaker.Do(test.req)
					if err != nil {
						fmt.Printf("%v\n", err)
					}
					fmt.Println()
				}()
			}
		})
	}
}

func TestBreakerDemo(t *testing.T) {
	// complexBreaker
	err := NewBreaker().
		WithAcceptable(func(err error) bool {
			return true
		}).
		WithFallback(func(err error) error {
			return nil
		}).
		WithBreakerType(define.GoogleBreaker).
		GetOrBuild("123").
		Do(func() error {
			fmt.Printf("123")
			return nil
		})
	// simpleBreaker
	err = NewBreaker().
		Do(func() error {
			return nil
		})
	fmt.Printf("%v", err)
}
