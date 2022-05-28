package consumer

import (
	"errors"
	"fmt"
	"sync"
)

var mg Manager
var once sync.Once

func init() {
	once.Do(func() {
		mg = Manager{
			instances: make(map[string]*Consumer),
		}
	},
	)
}

type Manager struct {
	instances map[string]*Consumer
	mutex     sync.Mutex
}

//注册实例
func Register(instance string, setting *setting) error {
	defer mg.mutex.Unlock()
	mg.mutex.Lock()

	if _, ok := mg.instances[instance]; ok {
		return errors.New(fmt.Sprintf("instance:%s has exists!", instance))
	}
	cf, err := NewConsumer(setting)
	if err != nil {
		return err
	}
	mg.instances[instance] = cf
	return nil
}

func GetConsumer(instance string) (*Consumer, error) {
	if ins, ok := mg.instances[instance]; ok {
		return ins, nil
	} else {
		return nil, errors.New(fmt.Sprintf("instance:%s not exists!", instance))
	}
}
func Reset() {
	defer mg.mutex.Unlock()
	mg.mutex.Lock()
	for _, item := range mg.instances {
		item.Close()
	}
	mg.instances = make(map[string]*Consumer)
}
