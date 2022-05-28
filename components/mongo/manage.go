package mongo

import (
	"errors"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

var mg Manager
var once sync.Once

func init() {
	once.Do(func() {
		mg = Manager{
			instances: make(map[string]*Mongo),
		}
	},
	)
}

type Manager struct {
	instances map[string]*Mongo
	mutex     sync.Mutex
}

//注册实例
func Register(instance string, setting Setting) error {
	defer mg.mutex.Unlock()
	mg.mutex.Lock()

	if _, ok := mg.instances[instance]; ok {
		return errors.New(fmt.Sprintf("instance:%s has exists!", instance))
	}
	cf, err := NewMongo(setting)
	if err != nil {
		return err
	}
	mg.instances[instance] = cf
	return nil
}

func GetMongo(instance string) (*Mongo, error) {
	if ins, ok := mg.instances[instance]; ok {
		return ins, nil
	} else {
		return nil, errors.New(fmt.Sprintf("instance:%s not exists!", instance))
	}
}

func GetMongoClient(instance string) *mongo.Client {
	if ins, ok := mg.instances[instance]; ok {
		return ins.Client
	}
	return nil
}

func Reset() {
	defer mg.mutex.Unlock()
	mg.mutex.Lock()
	for _, item := range mg.instances {
		item.Close()
	}
	mg.instances = make(map[string]*Mongo)
}
