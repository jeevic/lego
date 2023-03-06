package log

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jeevic/lego/util"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Log struct {
	Logger *zerolog.Logger
	Closer func() error
	Writer io.Writer
}

func NewZeroLog(opts *Options) (*Log, error) {
	closers := make([]func() error, 0)
	writers := make([]io.Writer, 0)

	closer := func() error {
		var errs []string
		for _, closeFn := range closers {
			err := closeFn()
			if err != nil {
				errs = append(errs, err.Error())
			}
		}

		if len(errs) > 0 {
			return fmt.Errorf("%d logger close errors: %s",
				len(errs), strings.Join(errs, ". Also, "))
		}

		return nil
	}

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"
	if opts.ConsoleLoggingEnabled {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stdout})
	}

	if opts.Rotate && len(opts.Filename) > 0 {
		filename := opts.Directory + opts.Filename

		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, opts.FileMode)
		if err != nil {
			_ = closer()
			return &Log{}, err
		}

		// but then close it for rotator to rotate if necessary
		_ = f.Close()

		rotator := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    opts.RotateMaxSize,
			MaxAge:     int(opts.RotateKeepAge.Hours() / 24),
			MaxBackups: opts.RotateKeepNumber,
			Compress:   opts.RotateCompress,
		}

		writers = append(writers, rotator)
		closers = append(closers, func() error { return rotator.Close() })
	}

	if !opts.Rotate && len(opts.Filename) > 0 {
		filename := opts.Directory + opts.Filename
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, opts.FileMode)
		if err != nil {
			_ = closer()
			return &Log{}, err
		}

		writers = append(writers, f)
		closers = append(closers, func() error { return f.Close() })
	}

	//设置全局的日志level
	level, err := zerolog.ParseLevel(opts.Level)
	if err == nil {
		zerolog.SetGlobalLevel(level)
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	mw := zerolog.MultiLevelWriter(writers...)
	c := zerolog.New(mw).With().Timestamp()
	if opts.ReportHostIp {
		ip, _ := util.GetLocalIp()
		c = c.Str("host", ip)
	}

	if opts.ReportCaller {
		c = c.Caller()
	}

	l := c.Logger()
	return &Log{
		Logger: &l,
		Closer: closer,
		Writer: mw,
	}, nil
}
