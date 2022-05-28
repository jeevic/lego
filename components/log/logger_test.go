package log

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

//初始化logger
func initLogger(t *testing.T) *Log {
	dir, _ := os.Getwd()
	var path = dir + "/../../test/logs/"
	c := Setting{
		Path:            path,
		FileName:        "app.log",
		Level:           "trace",
		Split:           ".%Y%m%d%H",
		LifeTime:        1,
		Rotation:        1,
		Format:          "ydLog",
		ReportHostIp:    true,
		ReportShortFile: true,
		ReportCaller:    true,
	}
	logger, err := NewLog(c)
	assert.NotEqual(t, logger, nil, "log init error", err)

	return logger
}

func TestNewLog_WithNoSetting(t *testing.T) {
	c := Setting{}
	logger, _ := NewLog(c)
	assert.NotEqual(t, logger, nil, "logger eq nil")
	assert.Equal(t, logger.Writer, os.Stdout, "logger init os.stdout")
}

func TestLog_Debug(t *testing.T) {
	logger := initLogger(t)
	logger.Logger.Debug("debug")
}

func TestLog_Info(t *testing.T) {
	logger := initLogger(t)
	logger.Logger.Info("info")
}
