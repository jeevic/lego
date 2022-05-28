package grpcserver

import (
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type GrpcServer struct {
	option *Options
	//address
	target string
	Server *grpc.Server
}

func NewGrpcServer(target string, options ...Option) (*GrpcServer, error) {
	opts := NewOptions(options...)

	srvOptions := make([]grpc.ServerOption, 0, 5)

	//keepalive Enforcement
	var kaep = keepalive.EnforcementPolicy{}
	if opts.KeepaliveEnforcementPolicyMinTime > 0 {
		kaep.MinTime = opts.KeepaliveEnforcementPolicyMinTime
	}
	kaep.PermitWithoutStream = opts.KeepaliveEnforcementPolicyPermitWithoutStream
	srvOptions = append(srvOptions, grpc.KeepaliveEnforcementPolicy(kaep))

	//keepalive
	var kasp = keepalive.ServerParameters{}
	if opts.KeepaliveMaxConnectionIdle > 0 {
		kasp.MaxConnectionIdle = opts.KeepaliveMaxConnectionIdle
	}
	if opts.KeepaliveMaxConnectionAge > 0 {
		kasp.MaxConnectionAge = opts.KeepaliveMaxConnectionAge
	}
	if opts.KeepaliveMaxConnectionAgeGrace > 0 {
		kasp.MaxConnectionAgeGrace = opts.KeepaliveMaxConnectionAgeGrace
	}
	if opts.KeepaliveTime > 0 {
		kasp.Time = opts.KeepaliveTime
	}
	if opts.KeepaliveTimeout > 0 {
		kasp.Timeout = opts.KeepaliveTimeout
	}
	srvOptions = append(srvOptions, grpc.KeepaliveParams(kasp))

	srvOptions = append(srvOptions, grpc.Creds(opts.Credentials))

	if len(opts.UnaryInterceptors) > 0 {
		srvOptions = append(srvOptions, grpc.ChainUnaryInterceptor(opts.UnaryInterceptors...))
	}
	if len(opts.StreamInterceptors) > 0 {
		srvOptions = append(srvOptions, grpc.ChainStreamInterceptor(opts.StreamInterceptors...))
	}

	s := &GrpcServer{}
	s.option = opts
	s.Server = grpc.NewServer(srvOptions...)
	s.target = target
	return s, nil

}

//register server
func (s *GrpcServer) RegisterService(f func(s *grpc.Server, srv interface{}), ss interface{}) {
	f(s.Server, ss)
}

func (s *GrpcServer) Run() error {
	//reflection for query api
	reflection.Register(s.Server)

	lis, err := net.Listen("tcp", s.target)
	if err != nil {
		return err
	}
	err = s.Server.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}

func (s *GrpcServer) RunAsync() {
	go func() {
		_ = s.Run()
	}()
}

func (s *GrpcServer) GracefulShutdown() {
	s.Server.GracefulStop()
}
