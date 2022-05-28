package crontab

import (
	"time"

	"github.com/go-co-op/gocron"
)

//@see https://github.com/go-co-op/gocron
// usage:
//	crontab := New()
//	crontab.AddTaskFunc(func(scheduler Scheduler) {
//		_, _ = scheduler.Every(1).Second().Do(func() {
//			fmt.Println("every 1  second done")
//		})
//	})
//
//	crontab.StartAsync()
//	time.Sleep(2 * time.Second)
//
//	crontab.Clear()
//	crontab.Stop()

type Scheduler = *gocron.Scheduler

type Crontab struct {
	Scheduler *gocron.Scheduler
}

func New() *Crontab {
	location, _ := time.LoadLocation("Asia/Shanghai")
	s := gocron.NewScheduler(location)
	return &Crontab{Scheduler: s}
}

//添加任务
func (c *Crontab) AddTaskFunc(callbacks ...func(scheduler Scheduler)) {
	for _, f := range callbacks {
		f(c.Scheduler)
	}
}

//异步启动 不阻塞当前进程
func (c *Crontab) StartAsync() {
	c.Scheduler.StartAsync()
}

//阻塞开始
func (c *Crontab) StartBlocking() {
	c.Scheduler.StartBlocking()
}

//根据tag删除job
func (c *Crontab) RemoveJobByTag(tag string) error {
	return c.Scheduler.RemoveJobByTag(tag)
}

//清理所有的任务
func (c *Crontab) Clear() {
	c.Scheduler.Clear()
}

//任务长度
func (c *Crontab) Len() int {
	return c.Scheduler.Len()
}

//停止任务
func (c *Crontab) Stop() {
	c.Scheduler.Stop()
}
