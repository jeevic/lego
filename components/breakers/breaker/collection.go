package breaker

import (
	"github.com/jeevic/lego/components/breakers/utils"
	"sync"
	"time"
)

type (
	// RollingWindowOption let callers customize the RollingWindow.
	RollingWindowOption func(rollingWindow *RollingWindow)

	// RollingWindow defines a rolling windows to calculate the events in buckets with time interval.
	RollingWindow struct {
		lock          sync.RWMutex
		size          int
		win           *windows
		interval      time.Duration
		offset        int
		ignoreCurrent bool
		lastTime      time.Duration // start time of the last bucket
	}
)

// NewRollingWindow returns a RollingWindow that with size buckets and time interval,
// use opts to customize the RollingWindow.
func NewRollingWindow(size int, interval time.Duration, opts ...RollingWindowOption) *RollingWindow {
	if size < 1 {
		panic("size must be greater than 0")
	}

	w := &RollingWindow{
		size:     size,
		win:      newWindow(size), // 包含 n个bucket，每个 bucket 记录 sum和count
		interval: interval,        //  每个窗口的时间间隔
		lastTime: utils.Now(),     // start time of the last bucket
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

// Add adds value to current bucket.
func (rw *RollingWindow) Add(v float64) {
	rw.lock.Lock()
	defer rw.lock.Unlock()
	// 计数的核心
	rw.updateOffset()
	rw.win.add(rw.offset, v)
}

// Reduce runs fn on all buckets, ignore current bucket if ignoreCurrent was set.
func (rw *RollingWindow) Reduce(fn func(b *Bucket)) {
	rw.lock.RLock()
	defer rw.lock.RUnlock()

	var diff int
	span := rw.span()
	// ignore current bucket, because of partial data
	if span == 0 && rw.ignoreCurrent {
		diff = rw.size - 1
	} else {
		diff = rw.size - span
	}
	if diff > 0 {
		offset := (rw.offset + span + 1) % rw.size
		rw.win.reduce(offset, diff, fn)
	}
}

func (rw *RollingWindow) span() int {
	offset := int(utils.Since(rw.lastTime) / rw.interval)
	if 0 <= offset && offset < rw.size {
		return offset
	}

	return rw.size
}

// 一开始在初始化breaker的时候，10S为单位分成40份
// 这里用到了一个巧妙方式自动初始化一部分区间的数值，也就是算法一开始所描述的，当错误次数过高熔断被触发，当一段时间过期后便会重新再去调用req方法,
// 只所以会过一段时间去调用，是因为sre公式又成立了，之所以成立了就是因为这里随着时间的推移初始化掉
// 了一部分区间的数值
func (rw *RollingWindow) updateOffset() {
	// breaker初始化过程中，|--|--|--|--|--|...   每个区间时间长度为: 10 * time.Second / 40，对应代码为数组的40个下标，
	// 不同时刻 闭包执行的统计值的累加结果存在不同下标的对象中进行累加, 这里的span()就是根据当前时刻决定下标移动的跨度
	span := rw.span()
	if span <= 0 {
		return
	}

	offset := rw.offset // 当前区间的下标
	// reset expired buckets, 重置过期的区间
	for i := 0; i < span; i++ {
		rw.win.resetBucket((offset + i + 1) % rw.size)
	}

	rw.offset = (offset + span) % rw.size
	now := utils.Now()
	// align to interval time boundary
	rw.lastTime = now - (now-rw.lastTime)%rw.interval
}

// Bucket defines the bucket that holds sum and num of additions.
type Bucket struct {
	Sum   float64
	Count int64
}

func (b *Bucket) add(v float64) {
	b.Sum += v
	b.Count++
}

func (b *Bucket) reset() {
	b.Sum = 0
	b.Count = 0
}

type windows struct {
	buckets []*Bucket
	size    int
}

func newWindow(size int) *windows {
	buckets := make([]*Bucket, size)
	for i := 0; i < size; i++ {
		buckets[i] = new(Bucket)
	}
	return &windows{
		buckets: buckets,
		size:    size,
	}
}

func (w *windows) add(offset int, v float64) {
	w.buckets[offset%w.size].add(v)
}

func (w *windows) reduce(start, count int, fn func(b *Bucket)) {
	for i := 0; i < count; i++ {
		fn(w.buckets[(start+i)%w.size])
	}
}

func (w *windows) resetBucket(offset int) {
	w.buckets[offset%w.size].reset()
}

// IgnoreCurrentBucket lets the Reduce call ignore current bucket.
func IgnoreCurrentBucket() RollingWindowOption {
	return func(w *RollingWindow) {
		w.ignoreCurrent = true
	}
}
