package interceptor

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"github.com/jeevic/lego/components/grpc/grpcserver"
	"github.com/jeevic/lego/pkg/app"
)

//this is a log unary or stream

func LogUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	requestId := ""
	clientIp := ""
	path := ""
	errMsg := ""

	requestId = grpcserver.FromContextRequestId(ctx, app.App.GetRequestId())
	p, ok := peer.FromContext(ctx)
	if ok {
		clientIp = p.Addr.String()
	}

	path = info.FullMethod
	start := time.Now()
	m, err := handler(ctx, req)
	if err != nil {
		errMsg = err.Error()
	}
	latency := time.Now().Sub(start)
	format := "grpc unary requestId=%s, client-ip=%s, path=%s, latency=%s, error-message=%s \n"
	app.App.GetLogger().Infof(format, requestId, clientIp, path, latency, errMsg)
	return m, err
}

func LogStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	requestId := ""
	clientIp := ""
	path := ""
	errMsg := ""
	requestId = grpcserver.FromContextRequestId(ss.Context(), app.App.GetRequestId())
	p, ok := peer.FromContext(ss.Context())
	if ok {
		clientIp = p.Addr.String()
	}
	path = info.FullMethod
	start := time.Now()
	err := handler(srv, ss)
	if err != nil {
		errMsg = err.Error()
	}
	latency := time.Now().Sub(start)
	format := "grpc stream res requestId=%s, client-ip=%s, path=%s, latency=%s, error-message=%s \n"
	app.App.GetLogger().Infof(format, requestId, clientIp, path, latency, errMsg)
	return err
}
