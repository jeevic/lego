package mongo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

//document @see https://godoc.org/go.mongodb.org/mongo-driver
//mongodb uri:  @see https://docs.mongodb.com/manual/reference/connection-string/
// usage:
//
//	hosts := "10.136.158.10:27000,10.136.158.10:27001,10.136.158.10:27002"
//	replset := "rs_image"
//
//	setting := Setting{
//		Hosts: hosts,
//		ReplSet: replset,
//	}
//	Mongo, err := NewMongo(&setting)
//	if err != nil {
//		return
//	}
//
//	c := Mongo.Client.Database("test").Collection("user")
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	cur, err := c.Find(ctx, bson.D{})
//	if err != nil {
//		t.Error("find collection error:", err)
//		return
//	}
//	defer cur.Close(ctx)
//	for cur.Next(ctx) {
//		var result bson.M
//		err := cur.Decode(&result)
//		if err != nil {
//		}
//	}

var readPreferenceMap = map[string]*readpref.ReadPref{
	"primary":            readpref.Primary(),
	"primaryPreferred":   readpref.PrimaryPreferred(),
	"secondary":          readpref.Secondary(),
	"secondaryPreferred": readpref.PrimaryPreferred(),
	"nearest":            readpref.Nearest(),
}

type Mongo struct {
	Client  *mongo.Client
	Setting *Setting
}

type Setting struct {
	//Uri 形式  @see https://docs.mongodb.com/manual/reference/connection-string/
	Uri string
	//此四个参数和Uri 互斥
	Hosts    string
	ReplSet  string
	Username string
	Password string
	//
	authSource string

	//max conn size default: 100
	MaxPoolSize uint64
	//min conn size
	MinPoolSize uint64
	//unit second
	MaxIdleTime int
	//primary (Default)
	//primaryPreferred
	//secondary
	//secondaryPreferred
	//nearest
	ReadPreference string

	WriteConcern *Wc
}

//控制写入安全的级别  分为应答式写入以及非应答式写入
//对于强一致性场景，建议w>1或者等于majority，以及journal为true，否则w=0
//在副本集的情形下，建议通过配置文件来修改w以及设置wtimeout，以避免由于某个节点挂起导致无法应答
//{ w: <value>, j: <boolean>, wtimeout: <number> }
//w:  w:1(应答式写入) 要求确认操作已经传播到指定的单个mongod实例或副本集主实例(缺省为1)
//但是对于尝试向已关闭的套接字写入或者网络故障会返回异常信息
//	w:>1(用于副本集环境)
//	该值用于设定写入节点的数目，包括主节点
//
//	"majority"(大多数)
//	适用于集群架构，要求写入操作已经传递到绝大多数投票节点以及主节点后进行应答
//
//	<tag set>
//	要求写入操作已经传递到指定tag标记副本集中的成员后进行应答
// j : 该选项要求确认写操作已经写入journal日志之后应答客户端(需要开启journal功能)
// wtimeout:
// 该选项指定一个时间限制,以防止写操作无限制被阻塞导致无法应答给客户端
// @see https://blog.csdn.net/leshami/article/details/52913705
// @see https://docs.mongodb.com/manual/reference/write-concern/
//WriteConcern
type Wc struct {
	//string or integer
	W interface{}
	J bool
	// unit second
	WTimeout int
}

//初始化数据
func NewMongo(setting Setting) (*Mongo, error) {
	opts := buildOptions(setting)
	cli, err := mongo.NewClient(opts)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("new mongodb error:%s", err.Error()))
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = cli.Connect(ctx)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("connect mongodb error:%s", err.Error()))
	}
	return &Mongo{Client: cli, Setting: &setting}, nil
}

func (m *Mongo) GetClient() *mongo.Client {
	return m.Client
}

func (m *Mongo) Close() {
	ctx := context.Background()
	_ = m.Client.Disconnect(ctx)
}

func buildOptions(setting Setting) *options.ClientOptions {
	opts := options.Client()
	//first uri
	if len(setting.Uri) > 0 {
		opts.ApplyURI(setting.Uri)
	} else {
		//hosts
		if len(setting.Hosts) > 0 {
			opts.SetHosts(strings.Split(setting.Hosts, ","))
		}
		//Username
		if len(setting.Username) > 0 {
			auth := options.Credential{
				Username: setting.Username,
				Password: setting.Password,
			}
			if len(setting.authSource) > 0 {
				auth.AuthSource = setting.authSource
			}
			opts.SetAuth(auth)
		}
		//replySet
		opts.SetReplicaSet(setting.ReplSet)
	}
	//maxsize
	if setting.MaxPoolSize > 0 {
		opts.SetMaxPoolSize(setting.MaxPoolSize)
	}
	//min pool size
	if setting.MinPoolSize > 0 {
		opts.SetMinPoolSize(setting.MinPoolSize)
	}
	if setting.MaxIdleTime > 0 {
		opts.SetMaxConnIdleTime(time.Duration(setting.MaxIdleTime) * time.Second)
	}
	//readPreference
	if len(setting.ReadPreference) > 0 {
		if v, ok := readPreferenceMap[setting.ReadPreference]; ok {
			opts.SetReadPreference(v)
		}
	}

	//write concern
	if setting.WriteConcern != nil {
		wOption := make([]writeconcern.Option, 0, 3)
		//w
		switch (setting.WriteConcern.W).(type) {
		case int:
			w := setting.WriteConcern.W.(int)
			wOption = append(wOption, writeconcern.W(w))
		case string:
			w := setting.WriteConcern.W.(string)
			if w == "majority" {
				wOption = append(wOption, writeconcern.WMajority())
			} else {
				wOption = append(wOption, writeconcern.WTagSet(w))
			}
		}
		//j
		wOption = append(wOption, writeconcern.J(setting.WriteConcern.J))
		//wTimeout
		if setting.WriteConcern.WTimeout > 0 {
			wOption = append(wOption, writeconcern.WTimeout(time.Second*time.Duration(setting.WriteConcern.WTimeout)))
		}
		opts.SetWriteConcern(writeconcern.New(wOption...))
	}
	return opts
}
