package grpcserver

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	defaultKeepaliveEnforcementPolicyMinTime             = 5 * time.Second
	defaultKeepaliveEnforcementPolicyPermitWithoutStream = true

	defaultKeepaliveMaxConnectionIdle = 120 * time.Second

	defaultKeepaliveTime    = 60 * time.Second
	defaultKeepaliveTimeOut = 10 * time.Second
)

// Option used by the Client
type Option func(*Options)

type Options struct {
	// MinTime is the minimum amount of time a client should wait before sending
	// a keepalive ping.
	KeepaliveEnforcementPolicyMinTime time.Duration
	// If true, server allows keepalive pings even when there are no active
	// streams(RPCs). If false, and client sends ping when there are no active
	// streams, server will send GOAWAY and close the connection.
	KeepaliveEnforcementPolicyPermitWithoutStream bool

	// MaxConnectionIdle is a duration for the amount of time after which an
	// idle connection would be closed by sending a GoAway. Idleness duration is
	// defined since the most recent time the number of outstanding RPCs became
	// zero or the connection establishment.
	//If a client is idle for 15 seconds, send a GOAWAY
	//// The current default value is infinity.
	KeepaliveMaxConnectionIdle time.Duration
	// If any connection is alive for more than 30 seconds, send a GOAWAY
	// MaxConnectionAge is a duration for the maximum amount of time a
	// connection may exist before it will be closed by sending a GoAway. A
	// random jitter of +/-10% will be added to MaxConnectionAge to spread out
	// connection storms.
	KeepaliveMaxConnectionAge time.Duration
	// Allow 5 seconds for pending RPCs to complete before forcibly closing connections
	// MaxConnectionAgeGrace is an additive period after MaxConnectionAge after
	// which the connection will be forcibly closed.
	KeepaliveMaxConnectionAgeGrace time.Duration
	// Ping the client if it is idle for 5 seconds to ensure the connection is still active
	// After a duration of this time if the server doesn't see any activity it
	// pings the client to see if the transport is still alive.
	// If set below 1s, a minimum value of 1s will be used instead.
	// The current default value is 2 hours.
	KeepaliveTime time.Duration

	// Wait 1 second for the ping ack before assuming the connection is dead
	// After having pinged for keepalive check, the server waits for a duration
	// of Timeout and if no activity is seen even after that the connection is
	// closed.
	// The current default value is 20 seconds.
	KeepaliveTimeout time.Duration

	//TransportCredentials  client use
	//@see https://github.com/grpc/grpc-go/tree/master/examples/features/encryption
	Credentials credentials.TransportCredentials

	//interceptor
	//@see https://github.com/grpc/grpc-go/tree/master/examples/features/interceptor
	UnaryInterceptors  []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor
}

func NewOptions(options ...Option) *Options {
	opts := NewDefaultOptions()
	for _, o := range options {
		o(opts)
	}
	return opts
}

//@see https://pkg.go.dev/google.golang.org/grpc/keepalive
func WithKeepaliveEnforcementPolicyMinTime(t time.Duration) Option {
	return func(o *Options) {
		o.KeepaliveEnforcementPolicyMinTime = t
	}
}

func WithKeepaliveEnforcementPolicyPermitWithoutStream(b bool) Option {
	return func(o *Options) {
		o.KeepaliveEnforcementPolicyPermitWithoutStream = b
	}
}

func WithKeepaliveMaxConnectionIdle(t time.Duration) Option {
	return func(o *Options) {
		o.KeepaliveMaxConnectionIdle = t
	}
}

func WithKeepaliveMaxConnectionAge(t time.Duration) Option {
	return func(o *Options) {
		o.KeepaliveMaxConnectionAge = t
	}
}

func WithKeepaliveMaxConnectionAgeGrace(t time.Duration) Option {
	return func(o *Options) {
		o.KeepaliveMaxConnectionAgeGrace = t
	}
}

func WithKeepaliveTime(t time.Duration) Option {
	return func(o *Options) {
		o.KeepaliveTime = t
	}
}

func WithKeepaliveTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.KeepaliveTimeout = t
	}
}

func WithCredentials(cred credentials.TransportCredentials) Option {
	return func(o *Options) {
		o.Credentials = cred
	}
}
func WithUnaryInterceptors(intercepters []grpc.UnaryServerInterceptor) Option {
	return func(o *Options) {
		o.UnaryInterceptors = intercepters
	}
}

func WithAppendUnaryInterceptor(intercepter grpc.UnaryServerInterceptor) Option {
	return func(o *Options) {
		if o.UnaryInterceptors == nil {
			o.UnaryInterceptors = make([]grpc.UnaryServerInterceptor, 0, 1)
		}
		o.UnaryInterceptors = append(o.UnaryInterceptors, intercepter)
	}
}

func WithStreamInterceptors(intercepters []grpc.StreamServerInterceptor) Option {
	return func(o *Options) {
		o.StreamInterceptors = intercepters
	}
}

func WithAppendStreamInterceptor(intercepter grpc.StreamServerInterceptor) Option {
	return func(o *Options) {
		if o.StreamInterceptors == nil {
			o.StreamInterceptors = make([]grpc.StreamServerInterceptor, 0, 1)
		}
		o.StreamInterceptors = append(o.StreamInterceptors, intercepter)
	}
}

func NewDefaultOptions() *Options {
	return &Options{
		KeepaliveEnforcementPolicyMinTime:             defaultKeepaliveEnforcementPolicyMinTime,
		KeepaliveEnforcementPolicyPermitWithoutStream: defaultKeepaliveEnforcementPolicyPermitWithoutStream,
		KeepaliveMaxConnectionIdle:                    defaultKeepaliveMaxConnectionIdle,

		// 默认false 如果有Credentials 设置为True
		KeepaliveTime:    defaultKeepaliveTime,
		KeepaliveTimeout: defaultKeepaliveTimeOut,
		Credentials:      insecure.NewCredentials(),
	}
}
