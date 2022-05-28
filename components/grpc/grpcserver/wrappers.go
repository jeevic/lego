package grpcserver

import (
	"context"

	"google.golang.org/grpc"
)

// WrappedServerStream is a thin wrapper around grpc.ServerStream that allows modifying context.
type WrappedServerStream struct {
	grpc.ServerStream
	// WrappedContext is the wrapper's own Context. You can assign it.
	WrappedContext context.Context
}

// Context returns the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *WrappedServerStream) Context() context.Context {
	return w.WrappedContext
}

// WrapServerStream returns a ServerStream that has the ability to overwrite context.
func NewWrappedServerStream(stream grpc.ServerStream, ctx context.Context) grpc.ServerStream {
	if existing, ok := stream.(*WrappedServerStream); ok {
		return existing
	}
	return &WrappedServerStream{
		ServerStream:   stream,
		WrappedContext: ctx,
	}
}

func FromContextRequestId(ctx context.Context, requestIdKey string) string {
	id, ok := ctx.Value(requestIdKey).(string)
	if !ok {
		return ""
	}
	return id
}
