package grpcclient

import (
	"time"

	"google.golang.org/grpc/credentials"
)

var (
	defaultClientPoolCap    int64 = 5
	defaultDialTimeout            = 5 * time.Second
	defaultKeepAlive              = 30 * time.Second
	defaultKeepAliveTimeout       = 10 * time.Second
)

// Option used by the Client
type Option func(*Options)

type Options struct {
	// client pool capacity
	PoolCap int64

	//grpc client dail time out
	DialTimeout time.Duration

	//Insecure
	Insecure bool

	//TransportCredentials  client use
	//@see https://github.com/grpc/grpc-go/tree/master/examples/features/encryption
	Credentials credentials.TransportCredentials

	// keepAlive
	// send pings every time duration if there is no activity
	KeepAlive time.Duration
	//wait time duration for ping ack before considering the connection dead
	KeepAliveTimeout time.Duration
	// send pings even without active streams
	KeepAlivePermitWithoutStream bool
}

func NewOptions(options ...Option) *Options {
	opts := NewDefaultOptions()
	for _, o := range options {
		o(opts)
	}
	return opts
}

func WithPoolCap(capacity int64) Option {
	return func(o *Options) {
		o.PoolCap = capacity
	}
}

func WithDailTimeOut(timeout time.Duration) Option {
	return func(o *Options) {
		o.DialTimeout = timeout
	}
}

func WithInsecure(b bool) Option {
	return func(o *Options) {
		o.Insecure = b
	}
}

/**
 * @see https://github.com/grpc/grpc-go/tree/master/examples/features/encryption
 *
 */
func WithCredentials(credentials credentials.TransportCredentials) Option {
	return func(o *Options) {
		o.Credentials = credentials
		o.Insecure = false
	}
}

func WithKeepAlive(ka time.Duration) Option {
	return func(o *Options) {
		o.KeepAlive = ka
	}
}

func WithKeepAliveTimeout(kat time.Duration) Option {
	return func(o *Options) {
		o.KeepAliveTimeout = kat
	}
}

func WithKeepAlivePermitWithoutStream(b bool) Option {
	return func(o *Options) {
		o.KeepAlivePermitWithoutStream = b
	}
}

func NewDefaultOptions() *Options {
	return &Options{
		PoolCap:     defaultClientPoolCap,
		DialTimeout: defaultDialTimeout,
		// 默认false 如果有Credentials 设置为True
		Insecure: true,

		KeepAlive:                    defaultKeepAlive,
		KeepAliveTimeout:             defaultKeepAliveTimeout,
		KeepAlivePermitWithoutStream: false,
	}
}
