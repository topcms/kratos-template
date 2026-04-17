package main

import (
	"flag"
	"os"

	infralogger "github.com/topcms/kratos-infra/logger"
	"github.com/topcms/kratos-template/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

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

func buildLogConfig(runtimeEnv string, c *conf.Log) infralogger.Config {
	cfg := infralogger.Config{
		Driver: "zap",
		Level:  "info",
		Format: "json",
		Output: "stdout",
		Caller: true,
	}

	if c != nil {
		if c.Driver != "" {
			cfg.Driver = c.Driver
		}
		if c.Level != "" {
			cfg.Level = c.Level
		}
		if c.Format != "" {
			cfg.Format = c.Format
		}
		if c.Output != "" {
			cfg.Output = c.Output
		}
		cfg.Caller = c.Caller
	}

	if runtimeEnv == "dev" {
		if c == nil || c.Level == "" {
			cfg.Level = "debug"
		}
		if c == nil || c.Format == "" {
			cfg.Format = "console"
		}
	}
	return cfg
}

func main() {
	flag.Parse()

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

	logger, loggerCleanup, err := infralogger.New(
		buildLogConfig(env, bc.Log),
		infralogger.ServiceMeta{
			ID:      id,
			Name:    Name,
			Version: Version,
			Env:     env,
		},
	)
	if err != nil {
		panic(err)
	}
	defer loggerCleanup()
	log.SetLogger(logger)

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Auth, bc.Registry, bc.Log, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
