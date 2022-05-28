package redis

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-redis/redis"
)

var mg Manager
var once sync.Once

func init() {
	once.Do(func() {
		mg = Manager{
			instances: make(map[string]*Redis),
		}
	},
	)
}

type Manager struct {
	instances map[string]*Redis
	mutex     sync.Mutex
}

//注册实例
func Register(instance string, setting *Setting) error {
	defer mg.mutex.Unlock()
	mg.mutex.Lock()

	if _, ok := mg.instances[instance]; ok {
		return errors.New(fmt.Sprintf("instance:%s has exists!", instance))
	}
	cf := NewRedisUniversal(setting)
	mg.instances[instance] = cf
	return nil
}

func GetRedis(instance string) (*Redis, error) {
	if ins, ok := mg.instances[instance]; ok {
		return ins, nil
	} else {
		return nil, errors.New(fmt.Sprintf("instance:%s not exists!", instance))
	}
}

func GetRedisClient(instance string) *redis.UniversalClient {
	if ins, ok := mg.instances[instance]; ok {
		return &ins.Client
	}
	return nil
}

func Reset() {
	defer mg.mutex.Unlock()
	mg.mutex.Lock()
	for _, item := range mg.instances {
		item.Close()
	}
	mg.instances = make(map[string]*Redis)
}
