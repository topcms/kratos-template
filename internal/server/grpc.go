package server

import (
	v1 "github.com/topcms/kratos-template/api/user/v1"
	"github.com/topcms/kratos-template/internal/conf"
	"github.com/topcms/kratos-template/internal/service"

	infralogging "github.com/topcms/kratos-infra/middleware/logging"
	infrarecovery "github.com/topcms/kratos-infra/middleware/recovery"
	infratracing "github.com/topcms/kratos-infra/middleware/tracing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, user *service.UserService, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			infratracing.Server(),
			infralogging.Server(logger),
			infrarecovery.Server(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout > 0 {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterUserServiceServer(srv, user)
	return srv
}
