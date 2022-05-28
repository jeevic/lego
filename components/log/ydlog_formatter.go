package log

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

/**
 *
 *
 */
type YdLogFormatter struct {
	// TimestampFormat - default: time.StampMilli = "Jan _2 15:04:05.000"
	TimestampFormat string
	//机器hostip
	HostIp string
	// print function caller recursion
	ReportCaller bool
	//报告report host ip
	ReportHostIp bool
	//报告短文件
	ReportShortFile bool
	// CustomCallerFormatter - set custom formatter for caller info  filename, line number
	CustomCallerFormatter func(*runtime.Frame) string
}

func (y *YdLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	//时间格式定义
	timestampFormat := y.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.StampMilli
	}
	//write time first
	b.WriteString(entry.Time.Format(timestampFormat))
	//write level
	y.appendKeyValue(b, "|", entry.Level.String())
	//主机 ip
	if y.ReportHostIp {
		y.appendKeyValue(b, "|", y.HostIp)
	}
	//写线程id
	y.appendKeyValue(b, "|", "0")
	//need write caller
	if y.ReportCaller {
		y.writeCaller(b, entry)
	}

	//write key value
	for k, v := range entry.Data {
		b.WriteByte('|')
		y.appendKeyValue(b, fmt.Sprintf("%s=", k), v)
	}
	// write mssage
	y.appendKeyValue(b, "|", entry.Message)
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (y *YdLogFormatter) writeCaller(b *bytes.Buffer, entry *logrus.Entry) {
	if entry.HasCaller() {
		if y.CustomCallerFormatter != nil {
			//自己的个性化
			b.WriteString(y.CustomCallerFormatter(entry.Caller))
			y.appendKeyValue(b, "|", y.CustomCallerFormatter(entry.Caller))
		} else {
			f := ""
			fn := ""
			if y.ReportShortFile {
				f = y.getShortFile(entry.Caller.File)
				fn = y.getShortFile(entry.Caller.Function)
			} else {
				f = entry.Caller.File
				fn = y.getShortFile(entry.Caller.Function)
			}
			file := fmt.Sprintf("%s:%d %s", f, entry.Caller.Line, fn)
			y.appendKeyValue(b, "|", file)
		}
	}
}

func (y *YdLogFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	/*if b.Len() > 0 {
		b.WriteByte(' ')
	}*/
	b.WriteString(key)
	//b.WriteByte('=')
	y.appendValue(b, value)
}

func (y *YdLogFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}
	b.WriteString(fmt.Sprintf("%s", stringVal))
}

func (y *YdLogFormatter) getShortFile(file string) string {
	s := strings.Split(file, "/")
	return s[len(s)-1]
}
