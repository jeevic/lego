package zookeeper

import (
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

//ZooKeeper builder
type ZkBuilder struct {
	Conn    *zk.Conn
	Setting *Setting
}

//config
type Setting struct {
	Hosts          []string
	SessionTimeout time.Duration
}

//实例化 zookeeper
func NewZkBuilder(hosts []string, t time.Duration) (*ZkBuilder, error) {
	zc := Setting{
		Hosts:          hosts,
		SessionTimeout: t,
	}
	zb := &ZkBuilder{
		Setting: &zc,
	}
	err := zb.Start()
	return zb, err
}

//start
func (zb *ZkBuilder) Start() error {
	conn, _, err := zk.Connect(zb.Setting.Hosts, zb.Setting.SessionTimeout)
	if err != nil {
		return err
	}
	zb.Conn = conn
	return nil
}

func (zb *ZkBuilder) Restart() error {
	zb.Conn.Close()
	return zb.Start()
}

//停止
func (zb *ZkBuilder) Stop() {
	zb.Conn.Close()
	zb.Conn = nil
	zb.Setting = nil
}

//创建节点
//path 节点路径 如：/contech/image-manager/test
//flag 0永久  1临时节点   2有序节点   3临时有序节点
func (zb *ZkBuilder) CreateZkNode(path string, flag int32) error {

	acl := zk.WorldACL(zk.PermAll)

	pathArr := strings.Split(path, "/")
	conn := zb.Conn
	basePath := ""

	for i := 0; i < len(pathArr); i++ {
		if len(pathArr[i]) <= 0 {
			continue
		}
		basePath = basePath + "/" + pathArr[i]
		exist, _, err := conn.Exists(basePath)
		if err != nil {
			return err
		}
		if !exist {
			_, err = conn.Create(basePath, nil, flag, acl)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

//更新zk节点数据
func (zb *ZkBuilder) UpdateZkData(path string, data []byte) error {
	conn := zb.Conn
	exist, stat, err := conn.Exists(path)
	if err != nil {
		return err
	}

	if !exist {
		err := zb.CreateZkNode(path, 0)
		if err != nil {
			return err
		}
	}
	version := stat.Version
	_, err = conn.Set(path, data, version)
	if err != nil {
		return err
	}
	return nil
}
