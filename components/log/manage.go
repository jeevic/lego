package log

import (
	"errors"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

var mg Manager
var once sync.Once

func init() {
	once.Do(func() {
		mg = Manager{
			instances: make(map[string]*Log),
		}
	},
	)
}

type Manager struct {
	instances map[string]*Log
	mutex     sync.Mutex
}

//注册实例
func Register(instance string, setting Setting) error {
	defer mg.mutex.Unlock()
	mg.mutex.Lock()

	if _, ok := mg.instances[instance]; ok {
		return errors.New(fmt.Sprintf("instance:%s has exists!", instance))
	}
	cf, err := NewLog(setting)
	if err != nil {
		return err
	}
	mg.instances[instance] = cf
	return nil
}

func GetLog(instance string) (*Log, error) {
	if ins, ok := mg.instances[instance]; ok {
		return ins, nil
	} else {
		return nil, errors.New(fmt.Sprintf("instance:%s not exists!", instance))
	}
}

func GetLogger(instance string) *logrus.Logger {
	if ins, ok := mg.instances[instance]; ok {
		return ins.Logger
	}
	return nil
}

func Reset() {
	defer mg.mutex.Unlock()
	mg.mutex.Lock()
	mg.instances = make(map[string]*Log)
}
