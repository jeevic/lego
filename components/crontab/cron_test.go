package crontab

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//添加任务方法
func TestCrontab_AddTaskFunc(t *testing.T) {
	crontab := New()
	var run int

	crontab.StartAsync()
	crontab.AddTaskFunc(func(scheduler Scheduler) {
		_, _ = scheduler.Every(1).Second().Do(func() {
			fmt.Println("every 1  second done")
			run = 1
		})
	})
	time.Sleep(4 * time.Second)

	assert.Equal(t, run, 1, "run must be 1")

	crontab.Clear()
	crontab.Stop()
}

//任务长度
func TestCrontab_Len(t *testing.T) {
	crontab := New()
	crontab.AddTaskFunc(func(scheduler Scheduler) {
		_, _ = scheduler.Every(1).Second().Do(func() {
			fmt.Println("every 1  second done")
		})
	})

	assert.Equal(t, crontab.Len(), 1, "crontab len 1")

	crontab.AddTaskFunc(func(scheduler Scheduler) {
		_, _ = scheduler.Every(2).Second().Do(func() {
			fmt.Println("every 1  second done")
		})
	})
	assert.Equal(t, crontab.Len(), 2, "crontab len 2")

	crontab.Clear()
	assert.Equal(t, crontab.Len(), 0, "crontab len 0")
}

//删除job
func TestCrontab_RemoveJobByTag(t *testing.T) {
	crontab := New()
	var tags = []string{"tag_name"}
	crontab.AddTaskFunc(func(scheduler Scheduler) {
		_, _ = scheduler.Every(1).Second().SetTag(tags).Do(func() {
			fmt.Println("every 1  second done")
		})
	})
	assert.Equal(t, crontab.Len(), 1, "crontab len 1")

	err := crontab.RemoveJobByTag(tags[0])
	assert.Equal(t, err, nil, "remove tag success")
}
