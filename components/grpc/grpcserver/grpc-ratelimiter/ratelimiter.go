package grpc_ratelimiter

import (
	"context"
	"sync"
	"time"

	"github.com/juju/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//全局限流标志
const LegoWholeAppSign = "_lego_whole_app#lego"

//限流模块
var RateLimiter = RateLimit{
	RB: make(map[string]*ratelimit.Bucket),
}

type RateLimit struct {
	//存储url对应的bucket
	RB    map[string]*ratelimit.Bucket
	mutex sync.Mutex
}

//添加对应的路径和限速速率
func (r *RateLimit) AddRateLimit(path string, capacity int64) {
	defer r.mutex.Unlock()
	r.mutex.Lock()
	r.RB[path] = ratelimit.NewBucketWithQuantum(1*time.Second, capacity, capacity)
}

//单容器进行限速 设置后 对单App不生效
func (r *RateLimit) AddWholeRateLimit(capacity int64) {
	defer r.mutex.Unlock()
	r.mutex.Lock()
	r.RB[LegoWholeAppSign] = ratelimit.NewBucketWithQuantum(1*time.Second, capacity, capacity)
}

// UnaryServerInterceptor returns a new unary server interceptors that performs request rate limiting.
func RateLimiterUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		path := info.FullMethod

		var bucket *ratelimit.Bucket
		var ok bool

		bucket, ok = RateLimiter.RB[LegoWholeAppSign]
		// 如果没有全局限流 则获取单个接口限流
		if !ok {
			//获取限流
			bucket, ok = RateLimiter.RB[path]
		}
		if ok {
			take := bucket.TakeAvailable(1)
			//如果未获取到take中断停止
			if take == 0 {
				return nil, status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_ratelimit middleware, please retry later.", info.FullMethod)
			}
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor that performs rate limiting on the request.
func RateLimiterStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		path := info.FullMethod

		var bucket *ratelimit.Bucket
		var ok bool
		bucket, ok = RateLimiter.RB[LegoWholeAppSign]
		// 如果没有全局限流 则获取单个接口限流
		if !ok {
			//获取限流
			bucket, ok = RateLimiter.RB[path]
		}
		if ok {
			take := bucket.TakeAvailable(1)
			//如果未获取到take中断停止
			if take == 0 {
				return status.Errorf(codes.ResourceExhausted, "%s is rejected by grpc_ratelimit middleware, please retry later.", info.FullMethod)
			}
		}
		return handler(srv, stream)
	}
}
