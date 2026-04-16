//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final binary.

package main

import (
	"github.com/topcms/kratos-template/internal/biz"
	"github.com/topcms/kratos-template/internal/conf"
	"github.com/topcms/kratos-template/internal/data"
	templateRegistry "github.com/topcms/kratos-template/internal/registry"
	"github.com/topcms/kratos-template/internal/server"
	"github.com/topcms/kratos-template/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Auth, *conf.Registry, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		templateRegistry.ProviderSet,
		newApp,
	))
}
