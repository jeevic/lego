package log

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"

	"github.com/jeevic/lego/util"
)

//@see https://github.com/sirupsen/logrus
//usage:
//
//	dir, _ := os.Getwd()
//	var path = dir + "/../../test/logs/"
//	c := Setting{
//	Path:         path,
//	FileName:     "app.log",
//	Level:        "trace",
//	Split:        ".%Y%m%d%H",
//	LifeTime:     1,
//	Rotation:     1,
//	Format:       "ydLog",
//	ReportCaller: true,
//	}
//	logger, err := NewLog(c)
//	logger.getLogger().Debug("debug")

type Log struct {
	//日志配置
	Setting *Setting
	//初始化日志句柄
	Logger *logrus.Logger
	Writer io.Writer
}

//日志配置信息
type Setting struct {
	Path            string        //log dir
	FileName        string        // Log filename
	ErrFileName     string        //错误日志目录
	Level           string        // log level
	Format          string        // text json or ydLog
	Split           string        // file spilt  .%Y%m%d
	LifeTime        time.Duration // 保存时间 单位 h
	Rotation        time.Duration //分割时间  单位 h
	ReportCaller    bool          //是否打印调用栈位置 行号
	ReportHostIp    bool          //是否打印host ip
	ReportShortFile bool          //文件路径短写
}

//实例化Log
func NewLog(setting Setting) (*Log, error) {
	h, w, err := InitLogrus(&setting)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("log init logrus error err:%s", err.Error()))
	}
	return &Log{Setting: &setting, Logger: h, Writer: w}, nil
}

//进行初始化
func InitLogrus(c *Setting) (*logrus.Logger, io.Writer, error) {
	l := logrus.New()
	//如果未设置path filename 直接返回
	if c == nil || len(c.Path) == 0 {
		l.SetOutput(os.Stdout)
		return l, os.Stdout, nil
	}

	basePath := path.Join(c.Path, c.FileName)
	writer, err := rotatelogs.New(
		basePath+c.Split,
		rotatelogs.WithLinkName(basePath),                 //生成软连接, 指向最新日志文件
		rotatelogs.WithMaxAge(c.LifeTime*time.Hour),       //文件最大保存时间 单位:时间
		rotatelogs.WithRotationTime(c.Rotation*time.Hour), //文件切割时间时间
	)
	if err != nil || l == nil {
		log.Printf("failed to create rotatelogs err:%s", err)
		return nil, nil, err
	}
	//错误文件地址
	var errWriter io.Writer
	if len(c.ErrFileName) > 0 {
		errPath := path.Join(c.Path, c.ErrFileName)
		ew, err := rotatelogs.New(
			errPath+c.Split,
			rotatelogs.WithLinkName(errPath),                  //生成软连接, 指向最新日志文件
			rotatelogs.WithMaxAge(c.LifeTime*time.Hour),       //文件最大保存时间 单位:时间
			rotatelogs.WithRotationTime(c.Rotation*time.Hour), //文件切割时间时间
		)
		if err != nil {
			log.Printf("failed to create error rotatelogs err:%s", err)
			return nil, nil, err
		}
		errWriter = io.MultiWriter(writer, ew)
	} else {
		errWriter = writer
	}

	//设置日志级别
	switch c.Level {
	case "trace":
		l.SetLevel(logrus.TraceLevel)
		break
	case "debug":
		l.SetLevel(logrus.DebugLevel)
		break
	case "info":
		l.SetLevel(logrus.InfoLevel)
		break
	case "warn":
		l.SetLevel(logrus.WarnLevel)
		break
	case "error":
		l.SetLevel(logrus.ErrorLevel)
		break
	case "fatal":
		l.SetLevel(logrus.FatalLevel)
		break
	case "panic":
		l.SetLevel(logrus.PanicLevel)
		break
	default:
		l.SetLevel(logrus.InfoLevel)
	}

	//聚合文件地址
	hook := lfshook.NewHook(
		lfshook.WriterMap{
			logrus.DebugLevel: writer,
			logrus.InfoLevel:  writer,
			logrus.WarnLevel:  writer,
			logrus.ErrorLevel: errWriter,
			logrus.FatalLevel: errWriter,
			logrus.PanicLevel: errWriter,
		},
		&logrus.JSONFormatter{},
	)

	//判断是否有error
	if c.ReportCaller {
		l.SetReportCaller(c.ReportCaller)
	}

	switch c.Format {
	case "json":
		hook.SetFormatter(&logrus.JSONFormatter{})
		break
	case "text":
		hook.SetFormatter(&logrus.TextFormatter{})
		break
	case "ydLog":
		//Host Ip
		ip, _ := util.GetLocalIp()
		hook.SetFormatter(&YdLogFormatter{
			TimestampFormat: "2006-01-02 15:04:05.000",
			HostIp:          ip,
			ReportCaller:    c.ReportCaller,
			ReportHostIp:    c.ReportHostIp,
			ReportShortFile: c.ReportShortFile,
		})
		break
	default:
		hook.SetFormatter(&logrus.JSONFormatter{})
	}
	l.Hooks.Add(hook)
	//将logrus 指定到 dev
	if devW, err := getDevNullWriter(); err == nil {
		l.SetOutput(devW)
	}
	//自带log输出位置设置
	log.SetOutput(writer)

	return l, writer, nil
}

func (l *Log) GetLogger() *logrus.Logger {
	return l.Logger
}

func getDevNullWriter() (io.Writer, error) {
	src, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	return bufio.NewWriter(src), err
}
