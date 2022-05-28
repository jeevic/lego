package godis

import (
	"errors"
	"time"

	"github.com/go-redis/redis"
	"github.com/go-zookeeper/zk"
)

//usage：
//	zkAddr := []string{"10.126.173.11:2181","10.126.173.12:2181", "10.126.173.13:2181"}
//	zkDir := "/jodis/codis-caifeng_test"
//
//	pool, err := Create().SetZookeeperClient(zkAddr, zkDir, 3000).SetDb(5).SetPoolSize(10).Build()
//  cli, _ := pool.GetClient()
//	cmd := cli.Set("test_1", 10, 100 * time.Second)
//	str, err := cmd.Result()

type Builder struct {
	//zookeeper 相关
	zkAddr             []string
	zkProxyDir         string
	zkSessionTimeoutMs time.Duration

	//redis 相关配置
	options redis.Options
}

func Create() *Builder {
	b := &Builder{
		zkSessionTimeoutMs: time.Millisecond * 5000,

		options: redis.Options{
			PoolSize:     20,
			MinIdleConns: 10,
			DB:           0,
		},
	}
	return b
}

//设置Zk客户端连接
func (b *Builder) SetZookeeperClient(zkAddr []string, zkProxyDir string, timeoutMs int) *Builder {
	b.zkAddr = zkAddr
	b.zkProxyDir = zkProxyDir
	b.zkSessionTimeoutMs = time.Millisecond * time.Duration(timeoutMs)
	return b
}

//设置redis配置项
func (b *Builder) SetRedisOptions(options redis.Options) *Builder {
	b.options = options
	return b
}

//设置redis passwd
func (b *Builder) SetPasswd(passwd string) *Builder {
	b.options.Password = passwd
	return b
}

func (b *Builder) SetDb(db int) *Builder {
	b.options.DB = db
	return b
}

func (b *Builder) SetPoolSize(poolSize int) *Builder {
	b.options.PoolSize = poolSize
	return b
}

func (b *Builder) SetMinIdleConns(minIdleConns int) *Builder {
	b.options.MinIdleConns = minIdleConns
	return b
}

//构造
func (b *Builder) Build() (*RoundRobinPool, error) {
	//创建zookeeper客户端
	zkConn, err := b.createZookeeperClient()
	if err != nil {
		return nil, err
	}
	return NewRoundRobinPool(zkConn, b.zkProxyDir, b.options)
}

func (b *Builder) createZookeeperClient() (*zk.Conn, error) {
	conn, _, err := zk.Connect(b.zkAddr, b.zkSessionTimeoutMs)
	if err != nil {
		return nil, errors.New("create zookeeper client error")
	}
	return conn, nil
}
