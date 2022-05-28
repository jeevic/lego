package app

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/jeevic/lego/components/config"
	"github.com/jeevic/lego/components/grpc/grpcserver"
	"github.com/jeevic/lego/components/httpserver"
	"github.com/jeevic/lego/components/log"
)

const (
	_ uint8 = iota
	DEVELOP
	TEST
	PERF
	PROD
)

//多实例默认标志

var DefaultInstance = "app"
var multiInstanceSign = "multi"

//环境变量
var envName2Num = map[string]uint8{
	"develop": DEVELOP,
	"test":    TEST,
	"perf":    PERF,
	"prod":    PROD,
}

var envNum2Name = map[uint8]string{
	DEVELOP: "develop",
	TEST:    "test",
	PERF:    "perf",
	PROD:    "prod",
}

func EnvName2Num(name string) (uint8, error) {
	if num, ok := envName2Num[name]; ok {
		return num, nil
	}
	return 0, errors.New(fmt.Sprintf("name:%s not found num", name))
}

func EnvNum2Name(num uint8) (string, error) {
	if str, ok := envNum2Name[num]; ok {
		return str, nil
	}
	return "", errors.New(fmt.Sprintf("num:%d not found name", num))
}

func IsMultiInstance(sign string) bool {
	return sign == multiInstanceSign
}

//设置到App
var App *Application
var Once sync.Once

type Application struct {
	Name string
	//环境参数
	Env uint8
	//配置路径
	CfgFile string
	//
	RequestId string
	//时区设置
	TimeLocationCST *time.Location
	//组件配置
	Components *Components

	mutex *sync.Mutex
}

//支持的组件
type Components struct {
	//配置 - 核心级别
	config struct {
		handler *config.Config
		enable  bool
	}
	//http server
	httpserver struct {
		handler *httpserver.HttpServer
		enable  bool
	}
	//GrpcServer
	grpcserver struct {
		handler *grpcserver.GrpcServer
		enable  bool
	}
}

func init() {
	Once.Do(func() {
		App = &Application{
			Components: &Components{},
			mutex:      new(sync.Mutex),
		}
	})
}

func (a *Application) SetName(name string) {
	a.Name = name
}

func (a *Application) GetName() string {
	return a.Name
}

func (a *Application) SetTimeLocation(timeZone string) error {
	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return err
	}
	a.TimeLocationCST = location
	return nil
}

func (a *Application) GetTimeLocation() *time.Location {
	return a.TimeLocationCST
}

func (a *Application) SetEnv(env string) error {

	if n, err := EnvName2Num(env); err == nil {
		a.Env = n
		return nil
	}
	return errors.New(fmt.Sprintf("application set env not exists env:%s", env))
}

func (a *Application) GetEnv() uint8 {
	return a.Env
}

func (a *Application) GetEnvName() string {
	str, _ := EnvNum2Name(a.Env)
	return str
}

//是否是开发环境
func (a *Application) IsDevelop() bool {
	return a.GetEnv() == DEVELOP
}

//是否是测试环境
func (a *Application) IsTest() bool {
	return a.GetEnv() == TEST
}

func (a *Application) IsPerf() bool {
	return a.GetEnv() == PERF
}

func (a *Application) IsProd() bool {
	return a.GetEnv() == PROD
}

//配置文件
func (a *Application) SetCfgFile(cfgFile string) {
	a.CfgFile = cfgFile
}

//配置文件获取
func (a *Application) GetCfgFile() (string, error) {
	if len(a.CfgFile) < 1 {
		return "", errors.New("no config file")
	}
	return a.CfgFile, nil
}

func (a *Application) SetRequestId(reqid string) {
	a.RequestId = reqid
}

func (a *Application) GetRequestId() string {
	return a.RequestId
}

//config
func (a *Application) SetConfig(cf *config.Config) {
	a.Components.config = struct {
		handler *config.Config
		enable  bool
	}{handler: cf, enable: true}
}

func (a *Application) GetConfig() (cf *config.Config, err error) {
	if a.Components.config.enable == false {
		return nil, errors.New("not init config")
	}
	return a.Components.config.handler, nil
}

func (a *Application) GetConfiger() *viper.Viper {
	cfg, _ := a.GetConfig()
	return cfg.Handler
}

func (a *Application) GetLog() (*log.Log, error) {
	hd, ok := log.GetLog(DefaultInstance)
	if ok != nil {
		return nil, errors.New("log not exists")
	}
	return hd, nil
}

func (a *Application) GetLogger() *logrus.Logger {
	l, _ := a.GetLog()
	return l.Logger
}

//httpserver
func (a *Application) SetHttpServer(hs *httpserver.HttpServer) {
	a.Components.httpserver = struct {
		handler *httpserver.HttpServer
		enable  bool
	}{handler: hs, enable: true}
}

func (a *Application) GetHttpServer() (*httpserver.HttpServer, error) {
	if a.Components.httpserver.enable == false {
		return nil, errors.New("not init httpserver")
	}
	return a.Components.httpserver.handler, nil
}

//grpc server
func (a *Application) SetGrpcServer(gs *grpcserver.GrpcServer) {
	a.Components.grpcserver = struct {
		handler *grpcserver.GrpcServer
		enable  bool
	}{handler: gs, enable: true}
}

func (a *Application) GetGrpcServer() (*grpcserver.GrpcServer, error) {
	if a.Components.grpcserver.enable == false {
		return nil, errors.New("not init grpc server")
	}
	return a.Components.grpcserver.handler, nil
}

func (a *Application) Close() {
	a.Components = &Components{}
}
