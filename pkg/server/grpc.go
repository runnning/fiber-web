package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type GrpcServer struct {
	server *grpc.Server
	config *GrpcConfig
	lis    net.Listener
}

type GrpcConfig struct {
	Addr                         string
	ReadTimeout                  time.Duration
	WriteTimeout                 time.Duration
	IdleTimeout                  time.Duration
	MaxRecvMsgSize               int
	MaxSendMsgSize               int
	InitialWindowSize            int32
	InitialConnWindowSize        int32
	MaxConcurrentStreams         uint32
	KeepAliveTime                time.Duration
	KeepAliveTimeout             time.Duration
	KeepAlivePermitWithoutStream bool
	EnableReflection             bool
	DisableStartupMessage        bool
	UnaryInterceptors            []grpc.UnaryServerInterceptor
	StreamInterceptors           []grpc.StreamServerInterceptor
}

type GrpcOption func(*GrpcConfig)

func defaultGrpcConfig() *GrpcConfig {
	return &GrpcConfig{
		Addr:                         ":50051",
		ReadTimeout:                  time.Second * 30,
		WriteTimeout:                 time.Second * 30,
		IdleTimeout:                  time.Second * 30,
		MaxRecvMsgSize:               4 * 1024 * 1024, // 4MB
		MaxSendMsgSize:               4 * 1024 * 1024, // 4MB
		InitialWindowSize:            32 * 1024,       // 32KB
		InitialConnWindowSize:        64 * 1024,       // 64KB
		MaxConcurrentStreams:         1000,
		KeepAliveTime:                time.Minute * 5,
		KeepAliveTimeout:             time.Second * 20,
		KeepAlivePermitWithoutStream: false,
		EnableReflection:             true,
		DisableStartupMessage:        false,
		UnaryInterceptors:            []grpc.UnaryServerInterceptor{},
		StreamInterceptors:           []grpc.StreamServerInterceptor{},
	}
}

func WithGrpcAddr(addr string) GrpcOption {
	return func(c *GrpcConfig) {
		c.Addr = addr
	}
}

func WithGrpcReadTimeout(t time.Duration) GrpcOption {
	return func(c *GrpcConfig) {
		c.ReadTimeout = t
	}
}

func WithGrpcWriteTimeout(t time.Duration) GrpcOption {
	return func(c *GrpcConfig) {
		c.WriteTimeout = t
	}
}

func WithGrpcIdleTimeout(t time.Duration) GrpcOption {
	return func(c *GrpcConfig) {
		c.IdleTimeout = t
	}
}

func WithGrpcMaxRecvMsgSize(size int) GrpcOption {
	return func(c *GrpcConfig) {
		c.MaxRecvMsgSize = size
	}
}

func WithGrpcMaxSendMsgSize(size int) GrpcOption {
	return func(c *GrpcConfig) {
		c.MaxSendMsgSize = size
	}
}

func WithGrpcMaxConcurrentStreams(n uint32) GrpcOption {
	return func(c *GrpcConfig) {
		c.MaxConcurrentStreams = n
	}
}

func WithGrpcKeepAlive(keepAliveTime, keepAliveTimeout time.Duration, permitWithoutStream bool) GrpcOption {
	return func(c *GrpcConfig) {
		c.KeepAliveTime = keepAliveTime
		c.KeepAliveTimeout = keepAliveTimeout
		c.KeepAlivePermitWithoutStream = permitWithoutStream
	}
}

func WithGrpcReflection(enable bool) GrpcOption {
	return func(c *GrpcConfig) {
		c.EnableReflection = enable
	}
}

func WithGrpcStartupMessage(disable bool) GrpcOption {
	return func(c *GrpcConfig) {
		c.DisableStartupMessage = disable
	}
}

func WithGrpcUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) GrpcOption {
	return func(c *GrpcConfig) {
		c.UnaryInterceptors = append(c.UnaryInterceptors, interceptors...)
	}
}

func WithGrpcStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) GrpcOption {
	return func(c *GrpcConfig) {
		c.StreamInterceptors = append(c.StreamInterceptors, interceptors...)
	}
}

func NewGrpcServer(opts ...GrpcOption) *GrpcServer {
	// 使用默认配置
	config := defaultGrpcConfig()

	// 应用所有选项
	for _, opt := range opts {
		opt(config)
	}

	// 创建 gRPC 服务器选项
	serverOptions := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(config.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(config.MaxSendMsgSize),
		grpc.InitialWindowSize(config.InitialWindowSize),
		grpc.InitialConnWindowSize(config.InitialConnWindowSize),
		grpc.MaxConcurrentStreams(config.MaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    config.KeepAliveTime,
			Timeout: config.KeepAliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             config.KeepAliveTime,
			PermitWithoutStream: config.KeepAlivePermitWithoutStream,
		}),
	}

	// 添加拦截器
	if len(config.UnaryInterceptors) > 0 {
		serverOptions = append(serverOptions, grpc.ChainUnaryInterceptor(config.UnaryInterceptors...))
	}
	if len(config.StreamInterceptors) > 0 {
		serverOptions = append(serverOptions, grpc.ChainStreamInterceptor(config.StreamInterceptors...))
	}

	// 创建 gRPC 服务器
	server := grpc.NewServer(serverOptions...)

	return &GrpcServer{
		server: server,
		config: config,
	}
}

func (s *GrpcServer) Server() *grpc.Server {
	return s.server
}

func (s *GrpcServer) Start() error {
	// 创建监听器
	lis, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	s.lis = lis

	// 打印启动信息
	if !s.config.DisableStartupMessage {
		fmt.Printf("\n[gRPC] Server listening on %s\n\n", s.config.Addr)
	}

	// 启动服务器
	return s.server.Serve(lis)
}

func (s *GrpcServer) Shutdown(ctx context.Context) error {
	// 使用 context 控制关闭超时
	done := make(chan error, 1)
	go func() {
		s.server.GracefulStop()
		if s.lis != nil {
			s.lis.Close()
		}
		done <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// RegisterService registers a gRPC service implementation to the server.
// Example usage:
// pb.RegisterUserServiceServer(s.Server(), &userService{})
func (s *GrpcServer) RegisterService(f func(srv *grpc.Server)) {
	f(s.server)
}
