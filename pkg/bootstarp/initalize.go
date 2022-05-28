package bootstarp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/credentials"

	"github.com/jeevic/lego/components/config"
	"github.com/jeevic/lego/components/grpc/grpcserver"
	grpc_ratelimiter "github.com/jeevic/lego/components/grpc/grpcserver/grpc-ratelimiter"
	"github.com/jeevic/lego/components/grpc/grpcserver/interceptor"
	"github.com/jeevic/lego/components/httpserver"
	"github.com/jeevic/lego/components/httpserver/middleware"
	"github.com/jeevic/lego/components/httpserver/ratelimiter"
	"github.com/jeevic/lego/components/log"
	"github.com/jeevic/lego/components/pprof"
	sig "github.com/jeevic/lego/components/signal"
	"github.com/jeevic/lego/components/swagger"
	"github.com/jeevic/lego/pkg/app"
	"github.com/jeevic/lego/util"
)

var initFunc = []func(){
	InitConfig,
	InitLog,
	InitApp,
	InitPid,
	InitHttpServer,
	InitGrpcServer,
	InitSwagger,
}

//初始化函数
func Init() error {
	t1 := time.Now()
	for _, f := range initFunc {
		f()
	}
	//注册信号函数
	sig.WatchSignal(func() {
		Stop(true)
	}, nil, nil)

	cost := time.Since(t1)
	app.App.GetLogger().Info("app init complete! time timeline:", cost)
	return nil
}

//注册初始化函数
func RegisterInit(f func()) {
	initFunc = append(initFunc, f)
}

//注册http route
func RegisterHttpRoutes(f func(engine *gin.Engine)) error {
	hs, _ := app.App.GetHttpServer()
	if hs == nil {
		return errors.New("http server not init")
	}
	f(hs.Engine)
	return nil
}

//注册信号监听函数
func RegisterSignalFunc(f sig.CallbackSignal) {
	sig.AddWatchFunc(f)
}

//初始化配置
func InitConfig() {
	cfg, err := app.App.GetCfgFile()
	if err != nil {
		panic(err.Error())
	}

	c, err := config.NewConfig(cfg)
	if err != nil {
		panic(fmt.Sprintf("[init] config error:%s", err.Error()))
	}
	//这是自动热加载文件
	c.WatchReConfig()
	app.App.SetConfig(c)
}

//初始化日志 -- 核心加载
//TODO 是否可以改成懒加载
func InitLog() {
	cfg := app.App.GetConfiger()
	//多实例
	if cfg.IsSet("log.type") && app.IsMultiInstance(cfg.GetString("log.type")) {
		instances := cfg.GetStringMap("log.instance")
		for instance := range instances {
			prefix := "log.instance." + instance + "."
			setting := log.Setting{
				Path:            cfg.GetString(prefix + "path"),
				FileName:        cfg.GetString(prefix + "filename"),
				ErrFileName:     cfg.GetString(prefix + "errfilename"),
				Level:           cfg.GetString(prefix + "level"),
				Format:          cfg.GetString(prefix + "format"),
				Split:           cfg.GetString(prefix + "split"),
				LifeTime:        cfg.GetDuration(prefix + "lifetime"),
				Rotation:        cfg.GetDuration(prefix + "rotation"),
				ReportCaller:    true,
				ReportHostIp:    true,
				ReportShortFile: true,
			}
			err := log.Register(instance, setting)
			if err != nil {
				panic(fmt.Sprintf("[init] log instance: %s error:%s", instance, err.Error()))
			}
		}
	} else {
		setting := log.Setting{
			Path:            cfg.GetString("log.path"),
			FileName:        cfg.GetString("log.filename"),
			ErrFileName:     cfg.GetString("log.errfilename"),
			Level:           cfg.GetString("log.level"),
			Format:          cfg.GetString("log.format"),
			Split:           cfg.GetString("log.split"),
			LifeTime:        cfg.GetDuration("log.lifetime"),
			Rotation:        cfg.GetDuration("log.rotation"),
			ReportCaller:    true,
			ReportHostIp:    true,
			ReportShortFile: true,
		}
		err := log.Register(app.DefaultInstance, setting)
		if err != nil {
			panic(fmt.Sprintf("[init] log error:%s", err.Error()))
		}
	}
	app.App.GetLogger().Info("[init] log component complete !")
}

//初始化app
func InitApp() {
	cfg := app.App.GetConfiger()
	name := cfg.GetString("app.name")
	app.App.SetName(name)
	app.App.GetLogger().Infof("[init] app name: %s !", name)

	//设置时区
	if cfg.IsSet("app.time_zone") {
		err := app.App.SetTimeLocation(cfg.GetString("app.time_zone"))
		if err != nil {
			_ = app.App.SetTimeLocation("Asia/Shanghai")
			app.App.GetLogger().Warn("[init] app init time zone error set default time zone")
		}
	}

	if cfg.IsSet("app.request_id") {
		app.App.SetRequestId(cfg.GetString("app.request_id"))
	}

	app.App.GetLogger().Info("[init] app init complete !")
}

//pid设置
func InitPid() {
	pid := os.Getpid()
	pidfile := app.App.GetConfiger().GetString("app.pidfile")
	if len(pidfile) < 1 {
		app.App.GetLogger().Infof("[init] not need init pid file")
		return
	}
	//判断当前pid 是否存储
	file, err := os.OpenFile(pidfile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		app.App.GetLogger().Warnf("[init] create pid file error:%s", err.Error())
		return
	}
	_, _ = file.WriteString(strconv.Itoa(pid))
	_ = file.Close()

	app.App.GetLogger().Infof("[init] create pid file pid:%d", pid)
}

func InitSwagger() {
	cfg := app.App.GetConfiger()
	enable := cfg.GetBool("swagger.enable")
	if enable {
		server, err := app.App.GetHttpServer()
		if err != nil && server != nil {
			swagger.InitSwagger(server.Engine)
			app.App.GetLogger().Infof("[init] swagger component complete!")
		}
	}
}

//初始化server
func InitHttpServer() {
	cfg := app.App.GetConfiger()
	if !cfg.IsSet("httpserver.http_host") {
		return
	}
	host := cfg.GetString("httpserver.http_host")
	port := cfg.GetInt("httpserver.http_port")
	isHttps := cfg.GetBool("httpserver.enable_https")
	middlewares := cfg.GetStringSlice("httpserver.middleware")

	//日志输出, 测试环境 双写
	l, _ := app.App.GetLog()
	outWriter := l.Writer
	if app.App.IsDevelop() {
		outWriter = io.MultiWriter(os.Stdout, outWriter)
	}
	//改写gin日志数据地址
	gin.DefaultErrorWriter = outWriter
	gin.DefaultWriter = outWriter

	hs := httpserver.NewHttpServer(host, port, isHttps)

	//非测试环境 打开
	if !app.App.IsDevelop() {
		hs.SetServerModeRelease()
	}

	//TODO 这段代码逻辑不太好
	if len(middlewares) > 0 {
		for _, mw := range middlewares {
			switch mw {
			case "cors":
				hs.SetMiddleware(middleware.CorsMiddleWare())
			case "requestid":
				hs.SetMiddleware(middleware.RequestIdMiddleware(app.App.GetRequestId()))
			case "ydlogger":
				//Host Ip
				ip, _ := util.GetLocalIp()
				hs.SetMiddleware(middleware.YdLoggerMiddleWare(outWriter, ip))
			case "pprof":
				pprof.UseHttpPprof(hs)
			}
		}
	}
	//设置限速模块
	hs.SetMiddleware(ratelimiter.RateLimitMiddleware())
	app.App.SetHttpServer(hs)
	app.App.GetLogger().Info("[init] http server complete!")
}

//grpc server
func InitGrpcServer() {
	cfg := app.App.GetConfiger()
	if !cfg.IsSet("grpcserver.grpc_host") {
		return
	}
	host := cfg.GetString("grpcserver.grpc_host")
	port := cfg.GetInt("grpcserver.grpc_port")

	options := make([]grpcserver.Option, 0, 10)

	//keepalive相关
	if cfg.IsSet("grpcserver.keepalive.enforcement_policy_mintime") {
		mintime := cfg.GetInt64("grpcserver.keepalive.enforcement_policy_mintime")
		if mintime > 0 {
			options = append(options, grpcserver.WithKeepaliveEnforcementPolicyMinTime(time.Duration(mintime)*time.Second))
		}
	}
	if cfg.IsSet("grpcserver.keepalive.enforcement_policy_permit_without_stream") {
		b := cfg.GetBool("grpcserver.keepalive.enforcement_policy_permit_without_stream")
		options = append(options, grpcserver.WithKeepaliveEnforcementPolicyPermitWithoutStream(b))
	}

	if cfg.IsSet("grpcserver.keepalive.max_connection_idle") {
		idle := cfg.GetInt64("grpcserver.keepalive.max_connection_idle")
		if idle > 0 {
			options = append(options, grpcserver.WithKeepaliveMaxConnectionIdle(time.Duration(idle)*time.Second))
		}
	}

	if cfg.IsSet("grpcserver.keepalive.max_connection_age") {
		age := cfg.GetInt64("grpcserver.keepalive.max_connection_age")
		if age > 0 {
			options = append(options, grpcserver.WithKeepaliveMaxConnectionAge(time.Duration(age)*time.Second))
		}
	}

	if cfg.IsSet("grpcserver.keepalive.max_connection_age_grace") {
		age_grace := cfg.GetInt64("grpcserver.keepalive.max_connection_age_grace")
		if age_grace > 0 {
			options = append(options, grpcserver.WithKeepaliveMaxConnectionAgeGrace(time.Duration(age_grace)*time.Second))
		}
	}

	if cfg.IsSet("grpcserver.keepalive.time") {
		time1 := cfg.GetInt64("grpcserver.keepalive.time")
		if time1 > 0 {
			options = append(options, grpcserver.WithKeepaliveTime(time.Duration(time1)*time.Second))
		}
	}

	if cfg.IsSet("grpcserver.keepalive.timeout") {
		timeout := cfg.GetInt64("grpcserver.keepalive.timeout")
		if timeout > 0 {
			options = append(options, grpcserver.WithKeepaliveTimeout(time.Duration(timeout)*time.Second))
		}
	}

	if cfg.IsSet("grpcserver.credentials") {
		server_cert := cfg.GetString("grpcserver.credentials.server_cert")
		server_key := cfg.GetString("grpcserver.credentials.server_key")

		creds, err := credentials.NewServerTLSFromFile(server_cert, server_key)
		if err != nil {
			app.App.GetLogger().Errorf("grpc credentials err:%v", err)
			panic(fmt.Sprintf("grpc credentials err:%v", err))
		}
		options = append(options, grpcserver.WithCredentials(creds))
	}

	//增加recovery
	options = append(options, grpcserver.WithAppendUnaryInterceptor(interceptor.DefaultRecoveryUnaryServerInterceptor()))
	options = append(options, grpcserver.WithAppendStreamInterceptor(interceptor.DefaultRecoveryStreamServerInterceptor()))

	if cfg.IsSet("grpcserver.interceptor") {
		interceptors := cfg.GetStringSlice("grpcserver.interceptor")
		//must first register this
		if b, _ := util.Contain("requestid", interceptors); b {
			options = append(options, grpcserver.WithAppendUnaryInterceptor(interceptor.RequestIdUnaryInterceptor))
			options = append(options, grpcserver.WithAppendStreamInterceptor(interceptor.RequestIdStreamInterceptor))
		}

		if b, _ := util.Contain("log", interceptors); b {
			options = append(options, grpcserver.WithAppendUnaryInterceptor(interceptor.LogUnaryInterceptor))
			options = append(options, grpcserver.WithAppendStreamInterceptor(interceptor.LogStreamInterceptor))
		}
	}
	//add ratelimiter
	options = append(options, grpcserver.WithAppendUnaryInterceptor(grpc_ratelimiter.RateLimiterUnaryServerInterceptor()))
	options = append(options, grpcserver.WithAppendStreamInterceptor(grpc_ratelimiter.RateLimiterStreamServerInterceptor()))

	target := fmt.Sprintf("%s:%d", host, port)

	gs, err := grpcserver.NewGrpcServer(target, options...)
	if err != nil {
		panic(err)
	}
	app.App.SetGrpcServer(gs)
	app.App.GetLogger().Info("[init] grpc server complete!")
}
