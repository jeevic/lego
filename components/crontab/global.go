package crontab

import "sync"

var globalCrontab *Crontab
var once sync.Once

func init() {
	once.Do(func() {
		globalCrontab = New()
		globalCrontab.StartAsync()
	})
}

//删除任务
func AddTaskFunc(callbacks ...func(scheduler Scheduler)) {
	globalCrontab.AddTaskFunc(callbacks...)
}

//删除tag
func RemoveJobByTag(tag string) error {
	return globalCrontab.RemoveJobByTag(tag)
}

//清理所有的任务
func Clear() {
	globalCrontab.Clear()
}

//任务长度
func Len() int {
	return globalCrontab.Len()
}

//停止任务
func Stop() {
	globalCrontab.Stop()
}
