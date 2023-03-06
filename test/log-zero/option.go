package log

import (
	"os"
	"time"
)

var (
	// DefaultFileMode is file Mode
	DefaultFileMode = os.FileMode(0644)
)

// Option used by the Client
type Option func(*Options)

// Options Config Configuration for logging
type Options struct {
	// Directory to log to to when filelogging is enabled
	Directory string

	// Filename is the name of the logfile which will be placed inside the directory
	Filename string

	//ErrFileName is the name of the error logfile which will be placed inside the directory
	ErrFileName string
	// FileMode to use when creating the file e.g. 0644, 0600
	FileMode os.FileMode

	// log level
	Level string

	// host ip输出
	ReportHostIp bool

	//是否打印调用栈位置 行号
	ReportCaller bool

	// Enable console logging
	ConsoleLoggingEnabled bool

	// If Rotate is true, logs are rotated (e.g. like logrotate) once they
	// get to a certain size.
	Rotate bool

	// If RotateCompress is true, rotated log files are compressed (read them
	// with zless, zcat, or gunzip, for example)
	RotateCompress bool

	// A log is rotated if it would be bigger than RotateMaxSize (in MB)
	RotateMaxSize int

	// If non-zero, delete any rotated logs older than RotateKeepAge
	RotateKeepAge time.Duration

	// If non-zero, keep only this many rotated logs and delete any exceeding
	// the limit of RotateKeepNumber.
	RotateKeepNumber int
}

func NewOptions(options ...Option) *Options {
	opts := NewDefaultOptions()
	for _, o := range options {
		o(opts)
	}
	return opts
}

func NewDefaultOptions() *Options {
	return &Options{
		ConsoleLoggingEnabled: true,
		FileMode:              DefaultFileMode,
		Rotate:                false,
	}
}

func WithDirectory(d string) Option {
	return func(options *Options) {
		options.Directory = d
	}
}

func WithFilename(filename string) Option {
	return func(options *Options) {
		options.Filename = filename
	}
}

func WithErrFileName(errFilename string) Option {
	return func(options *Options) {
		options.ErrFileName = errFilename
	}

}

func WithFileMode(mode int32) Option {
	return func(options *Options) {
		options.FileMode = os.FileMode(mode)
	}
}

func WithLevel(level string) Option {
	return func(options *Options) {
		options.Level = level
	}
}

func WithConsoleLoggingEnabled(b bool) Option {
	return func(options *Options) {
		options.ConsoleLoggingEnabled = b
	}
}

func WithRotate(rotate bool) Option {
	return func(options *Options) {
		options.Rotate = rotate
	}
}

func WithRotateCompress(compress bool) Option {
	return func(options *Options) {
		options.RotateCompress = compress
	}
}

func WithRotateMaxSize(maxsize int) Option {
	return func(options *Options) {
		options.RotateMaxSize = maxsize
	}
}

func WithRotateKeepAge(keepAge int) Option {
	return func(options *Options) {
		options.RotateKeepAge = time.Duration(keepAge) * time.Second
	}
}

func WithRotateKeepNumber(number int) Option {
	return func(options *Options) {
		options.RotateKeepNumber = number
	}
}

func WithReportHostIp(b bool) Option {
	return func(options *Options) {
		options.ReportHostIp = b
	}
}

func WithReportCaller(b bool) Option {
	return func(options *Options) {
		options.ReportCaller = b
	}
}
