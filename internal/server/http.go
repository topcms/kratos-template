package server

import (
	v1 "github.com/topcms/kratos-template/api/user/v1"
	"github.com/topcms/kratos-template/internal/conf"
	"github.com/topcms/kratos-template/internal/service"

	infraauth "github.com/topcms/kratos-infra/middleware/auth"
	infralogging "github.com/topcms/kratos-infra/middleware/logging"
	infrarecovery "github.com/topcms/kratos-infra/middleware/recovery"
	infratracing "github.com/topcms/kratos-infra/middleware/tracing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, user *service.UserService, logger log.Logger, validate infraauth.TokenValidator) *http.Server {
	mids := []middleware.Middleware{
		infratracing.Server(),
		infralogging.Server(logger),
	}
	if validate != nil {
		mids = append(mids, infraauth.Server(validate))
	}
	mids = append(mids, infrarecovery.Server())

	var opts = []http.ServerOption{
		http.Middleware(
			mids...,
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterUserServiceHTTPServer(srv, user)
	return srv
}
