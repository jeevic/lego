package zookeeper

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-zookeeper/zk"
)

var mg Manager
var once sync.Once

func init() {
	once.Do(func() {
		mg = Manager{
			instances: make(map[string]*ZkBuilder),
		}
	},
	)
}

type Manager struct {
	instances map[string]*ZkBuilder
	mutex     sync.Mutex
}

//注册实例
func Register(instance string, setting Setting) error {
	defer mg.mutex.Unlock()
	mg.mutex.Lock()

	if _, ok := mg.instances[instance]; ok {
		return errors.New(fmt.Sprintf("instance:%s has exists!", instance))
	}
	cf, err := NewZkBuilder(setting.Hosts, setting.SessionTimeout)
	if err != nil {
		return err
	}
	mg.instances[instance] = cf
	return nil
}

func GetZkBuilder(instance string) (*ZkBuilder, error) {
	if ins, ok := mg.instances[instance]; ok {
		return ins, nil
	} else {
		return nil, errors.New(fmt.Sprintf("instance:%s not exists!", instance))
	}
}

func GetZkBuilderConn(instance string) *zk.Conn {
	if ins, ok := mg.instances[instance]; ok {
		return ins.Conn
	}
	return nil
}

func Reset() {
	defer mg.mutex.Unlock()
	mg.mutex.Lock()
	for _, item := range mg.instances {
		item.Stop()
	}
	mg.instances = make(map[string]*ZkBuilder)
}
