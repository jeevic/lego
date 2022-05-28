package interceptor

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/jeevic/lego/components/grpc/grpcserver"
	"github.com/jeevic/lego/pkg/app"
)

func RequestIdUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var requestId = newRequestId()
	requestIdKey := app.App.GetRequestId()
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if val, ok := md[requestIdKey]; ok {
			if len(val) > 0 {
				requestId = val[0]
			}
		}
	}
	ctx = context.WithValue(ctx, requestIdKey, requestId)
	_ = grpc.SetHeader(ctx, metadata.MD{requestIdKey: []string{requestId}})
	m, err := handler(ctx, req)
	return m, err
}

func RequestIdStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	var requestId = newRequestId()
	requestIdKey := app.App.GetRequestId()
	ctx := ss.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if val, ok := md[requestIdKey]; ok {
			if len(val) > 0 {
				requestId = val[0]
			}
		}
	}
	ctx = context.WithValue(ctx, requestIdKey, requestId)
	ss = grpcserver.NewWrappedServerStream(ss, ctx)
	_ = ss.SetHeader(metadata.MD{requestIdKey: []string{requestId}})
	err := handler(srv, ss)
	return err
}

func newRequestId() string {
	return uuid.New().String()
}
