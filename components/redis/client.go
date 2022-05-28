package redis

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

// Universal redis client such as simple ,sentinel,cluster
// doc: https://github.com/go-redis/redis
type Redis struct {
	Client  redis.UniversalClient
	Setting *Setting
}

type Setting struct {
	MasterName string
	Hosts      []string
	//此四个参数和Uri 互斥
	Password string
	//max conn size default: 100
	MaxPoolSize int
	//min conn size
	MinPoolSize int
	//unit second
	MaxIdleTime int
	Db          int
	MaxRetries  int
}

func NewRedisUniversal(setting *Setting) *Redis {
	opts := buildUniversalOptions(setting)
	cli := redis.NewUniversalClient(opts)
	universalRedis := &Redis{Client: cli, Setting: setting}
	return universalRedis
}

func buildUniversalOptions(setting *Setting) *redis.UniversalOptions {
	options := &redis.UniversalOptions{}
	if len(setting.Hosts) > 0 {
		options.Addrs = setting.Hosts
	}
	if len(setting.Password) > 0 {
		options.Password = setting.Password
	}
	//maxsize
	if setting.MaxPoolSize > 0 {
		options.PoolSize = setting.MaxPoolSize
	}
	//min pool size
	if setting.MinPoolSize > 0 {
		options.MinIdleConns = setting.MinPoolSize
	}
	if setting.MaxIdleTime > 0 {
		options.IdleTimeout = time.Duration(setting.MaxIdleTime) * time.Second
	}
	if setting.Db >= 0 {
		options.DB = setting.Db
	}
	if len(setting.MasterName) > 0 {
		options.MasterName = setting.MasterName
	}
	if setting.MaxRetries > 0 {
		options.MaxRetries = setting.MaxRetries
	}
	return options
}

func (redis *Redis) Close() {
	err := redis.Client.Close()
	if err != nil {
		fmt.Println("close redis  error:" + err.Error())
	}
}
