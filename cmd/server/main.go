package main

import (
	"flag"
	"os"

	kratoszap "github.com/go-kratos/kratos/contrib/log/zap/v2"
	"github.com/topcms/kratos-template/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.uber.org/zap"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	Name string = "user.service"
	// Version is the version of the compiled software.
	Version  string
	flagconf string
	env      string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "configs", "config path (directory or file), eg: -conf configs")
	flag.StringVar(&Name, "name", Name, "service name used in registry/discovery")
	flag.StringVar(&env, "env", "prod", "runtime environment: dev or prod")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, registrar registry.Registrar) *kratos.App {
	opts := []kratos.Option{
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(gs, hs),
	}
	if registrar != nil {
		opts = append(opts, kratos.Registrar(registrar))
	}
	return kratos.New(opts...)
}

func newZapLogger(runtimeEnv string) (log.Logger, func(), error) {
	var zapCfg zap.Config
	if runtimeEnv == "dev" {
		zapCfg = zap.NewDevelopmentConfig() // console format
	} else {
		zapCfg = zap.NewProductionConfig() // json format
	}
	zl, err := zapCfg.Build()
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		_ = zl.Sync()
	}

	base := kratoszap.NewLogger(zl)
	logger := log.With(base,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
	)
	return logger, cleanup, nil
}

func main() {
	flag.Parse()
	logger, loggerCleanup, err := newZapLogger(env)
	if err != nil {
		panic(err)
	}
	defer loggerCleanup()
	log.SetLogger(logger)

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Auth, bc.Registry, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
