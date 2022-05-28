package sig

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

//signal 信号量处理工具
//usage:
// WatchSingal函数注册shutdown restart reconfig 三个函数即可

var sig *Signal
var once sync.Once

func init() {
	once.Do(func() {
		sig = &Signal{
			SignalChan: make(chan os.Signal, 1),
			Callbacks:  make([]CallbackSignal, 0),
			StopChan:   make(chan struct{}, 1),
		}
	})
}

type CallbackSignal func(sig os.Signal)

//信号量结构体
type Signal struct {
	SignalChan chan os.Signal
	StopChan   chan struct{}
	Callbacks  []CallbackSignal
	Running    bool
}

//addWatchFunc
func (s *Signal) AddWatchFunc(f CallbackSignal) *Signal {
	s.Callbacks = append(s.Callbacks, f)
	return s
}

func (s *Signal) WatchAsync() chan struct{} {
	//SIGINT SiGTERM  SIGHUP  终止信号
	//SIGUSR1 SIGUSR2 用户定义 SIGUSR1:定义平滑重启  SIGUSR2 配置重新加载
	signal.Notify(s.SignalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGHUP)
	s.Running = true
	go func() {
		for {
			sig := <-s.SignalChan
			var wg sync.WaitGroup
			for _, f := range s.Callbacks {
				f := f
				go func() {
					wg.Add(1)
					f(sig)
					wg.Done()
				}()
			}
			wg.Wait()

			if sig == syscall.SIGSTOP {
				s.Running = false
				//clear all watch func
				s.Clear()
				s.StopChan <- struct{}{}
				break
			}
		}

	}()

	return s.StopChan
}

func (s *Signal) Start() {
	<-s.WatchAsync()
}

func (s *Signal) StartAsync() {
	s.WatchAsync()
}

//only stop signal self
func (s *Signal) Stop() {
	if sig.Running {
		s.SignalChan <- syscall.SIGSTOP
	}
}

func (s *Signal) Clear() {
	s.Callbacks = make([]CallbackSignal, 0)
}

func AddWatchFunc(f CallbackSignal) {
	sig.AddWatchFunc(f)
}

func WatchSignal(shutdown func(), restart func(), reconfig func()) {
	callback := func(sig os.Signal) {
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP:
			//shutdown
			if shutdown != nil {
				shutdown()
			}
		case syscall.SIGUSR1:
			if restart != nil {
				restart()
			}
		case syscall.SIGUSR2:
			if reconfig != nil {
				reconfig()
			}
		}
	}
	sig.AddWatchFunc(callback)
	sig.StartAsync()
}

func Stop() {
	sig.Stop()
}
