package grpcclient

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var (
	ErrNotFoundClient = errors.New("not found grpc conn")
	ErrConnShutdown   = errors.New("grpc conn shutdown")
)

type GrpcClient struct {
	Conn *grpc.ClientConn
}

func NewClient(target string, options *Options) (*GrpcClient, error) {
	dopts := make([]grpc.DialOption, 0, 7)
	ctx := context.Background()
	var cancel context.CancelFunc
	if options.DialTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, options.DialTimeout)
		defer cancel()
	}

	if options.Insecure == true && options.Credentials == nil {
		dopts = append(dopts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	if options.Credentials != nil {
		dopts = append(dopts, grpc.WithTransportCredentials(options.Credentials))
	}

	// keepAlive
	if options.KeepAlive > 0 {
		dopts = append(dopts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                options.KeepAlive,
			Timeout:             options.KeepAliveTimeout,
			PermitWithoutStream: options.KeepAlivePermitWithoutStream,
		}))
	}

	conn, err := grpc.DialContext(ctx, target, dopts...)
	if err != nil {
		return nil, err
	}
	return &GrpcClient{conn}, nil
}

func (c *GrpcClient) Close() {
	if c.Conn != nil {
		_ = c.Conn.Close()
		c.Conn = nil
	}
}

//检验状态是不是关闭
func (c *GrpcClient) CheckState() error {
	if c.Conn == nil {
		return ErrNotFoundClient
	}
	state := c.Conn.GetState()
	switch state {
	case connectivity.TransientFailure, connectivity.Shutdown:
		return ErrConnShutdown
	}
	return nil
}

func (c *GrpcClient) GetConn() *grpc.ClientConn {
	return c.Conn
}
