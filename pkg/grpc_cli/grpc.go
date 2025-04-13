package grpc_cli

import (
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type GrpcClient struct {
	conn   *grpc.ClientConn
	config *GrpcConfig
}

type GrpcConfig struct {
	Addr                         string
	DialTimeout                  time.Duration
	MaxRecvMsgSize               int
	MaxSendMsgSize               int
	InitialWindowSize            int32
	InitialConnWindowSize        int32
	KeepAliveTime                time.Duration
	KeepAliveTimeout             time.Duration
	KeepAlivePermitWithoutStream bool
	DisableTLS                   bool
	UnaryInterceptors            []grpc.UnaryClientInterceptor
	StreamInterceptors           []grpc.StreamClientInterceptor
}

type GrpcOption func(*GrpcConfig)

func defaultGrpcConfig() *GrpcConfig {
	return &GrpcConfig{
		Addr:                         ":50051",
		DialTimeout:                  time.Second * 10,
		MaxRecvMsgSize:               4 * 1024 * 1024, // 4MB
		MaxSendMsgSize:               4 * 1024 * 1024, // 4MB
		InitialWindowSize:            32 * 1024,       // 32KB
		InitialConnWindowSize:        64 * 1024,       // 64KB
		KeepAliveTime:                time.Minute * 5,
		KeepAliveTimeout:             time.Second * 20,
		KeepAlivePermitWithoutStream: false,
		DisableTLS:                   true,
		UnaryInterceptors:            []grpc.UnaryClientInterceptor{},
		StreamInterceptors:           []grpc.StreamClientInterceptor{},
	}
}

func WithGrpcAddr(addr string) GrpcOption {
	return func(c *GrpcConfig) {
		c.Addr = addr
	}
}

func WithGrpcDialTimeout(t time.Duration) GrpcOption {
	return func(c *GrpcConfig) {
		c.DialTimeout = t
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

func WithGrpcKeepAlive(keepAliveTime, keepAliveTimeout time.Duration, permitWithoutStream bool) GrpcOption {
	return func(c *GrpcConfig) {
		c.KeepAliveTime = keepAliveTime
		c.KeepAliveTimeout = keepAliveTimeout
		c.KeepAlivePermitWithoutStream = permitWithoutStream
	}
}

func WithGrpcTLS(disable bool) GrpcOption {
	return func(c *GrpcConfig) {
		c.DisableTLS = disable
	}
}

func WithGrpcUnaryInterceptors(interceptors ...grpc.UnaryClientInterceptor) GrpcOption {
	return func(c *GrpcConfig) {
		c.UnaryInterceptors = append(c.UnaryInterceptors, interceptors...)
	}
}

func WithGrpcStreamInterceptors(interceptors ...grpc.StreamClientInterceptor) GrpcOption {
	return func(c *GrpcConfig) {
		c.StreamInterceptors = append(c.StreamInterceptors, interceptors...)
	}
}

func NewGrpcClient(opts ...GrpcOption) (*GrpcClient, error) {
	// 使用默认配置
	config := defaultGrpcConfig()

	// 应用所有选项
	for _, opt := range opts {
		opt(config)
	}

	// 创建 gRPC 客户端选项
	dialOptions := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(config.MaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(config.MaxSendMsgSize),
		),
		grpc.WithInitialWindowSize(config.InitialWindowSize),
		grpc.WithInitialConnWindowSize(config.InitialConnWindowSize),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                config.KeepAliveTime,
			Timeout:             config.KeepAliveTimeout,
			PermitWithoutStream: config.KeepAlivePermitWithoutStream,
		}),
		// 使用dns作为默认名称解析器
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}], "methodConfig": []}`),
	}

	// 添加TLS配置
	if config.DisableTLS {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// 添加拦截器
	if len(config.UnaryInterceptors) > 0 {
		dialOptions = append(dialOptions, grpc.WithChainUnaryInterceptor(config.UnaryInterceptors...))
	}
	if len(config.StreamInterceptors) > 0 {
		dialOptions = append(dialOptions, grpc.WithChainStreamInterceptor(config.StreamInterceptors...))
	}

	// 创建连接

	// 使用Connect替代NewClientConn
	conn, err := grpc.NewClient(config.Addr, dialOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return &GrpcClient{
		conn:   conn,
		config: config,
	}, nil
}

func (c *GrpcClient) Conn() *grpc.ClientConn {
	return c.conn
}

func (c *GrpcClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
