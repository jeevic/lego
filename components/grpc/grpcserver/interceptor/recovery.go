package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jeevic/lego/pkg/app"
)

var (
	RecoveryHandlerFunction = DefaultRecoveryHandlerFunc
)

// RecoveryHandlerFunc is a function that recovers from the panic `p` by returning an `error`.
type RecoveryHandlerFunc func(p interface{}) (err error)

// RecoveryHandlerFuncContext is a function that recovers from the panic `p` by returning an `error`.
// The context can be used to extract request scoped metadata and context values.
type RecoveryHandlerFuncContext func(ctx context.Context, p interface{}) (err error)

func DefaultRecoveryUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return RecoveryUnaryServerInterceptor(RecoveryHandlerFunction)
}

func DefaultRecoveryStreamServerInterceptor() grpc.StreamServerInterceptor {
	return RecoveryStreamServerInterceptor(RecoveryHandlerFunction)
}

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func RecoveryUnaryServerInterceptor(f RecoveryHandlerFuncContext) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		panicked := true

		defer func() {
			if r := recover(); r != nil || panicked {
				err = recoverFrom(ctx, r, f)
			}
		}()

		resp, err := handler(ctx, req)
		panicked = false
		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor for panic recovery.
func RecoveryStreamServerInterceptor(f RecoveryHandlerFuncContext) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		panicked := true

		defer func() {
			if r := recover(); r != nil || panicked {
				err = recoverFrom(stream.Context(), r, f)
			}
		}()

		err = handler(srv, stream)
		panicked = false
		return err
	}
}

func RegisterRecoveryHandler(f RecoveryHandlerFunc) {
	RecoveryHandlerFunction = RecoveryHandlerFuncContext(func(ctx context.Context, p interface{}) error {
		return f(p)
	})
}

func RecoveryHandlerContext(f RecoveryHandlerFuncContext) {
	RecoveryHandlerFunction = f
}

func DefaultRecoveryHandlerFunc(ctx context.Context, p interface{}) (err error) {
	app.App.GetLogger().Errorf("grpc server panic error:%v", p)
	return status.Errorf(codes.Internal, "%v", p)
}

func recoverFrom(ctx context.Context, p interface{}, r RecoveryHandlerFuncContext) error {
	if r == nil {
		return status.Errorf(codes.Internal, "%v", p)
	}
	return r(ctx, p)
}
