package registry

import (
	"fmt"

	infraConsul "github.com/topcms/kratos-infra/registry/consul"
	templateconf "github.com/topcms/kratos-template/internal/conf"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
)

// ProviderSet is registry providers.
var ProviderSet = wire.NewSet(
	NewConsulParts,
	NewRegistrarFromParts,
	NewDiscoveryFromParts,
)

type ConsulParts struct {
	Registrar registry.Registrar
	Discovery registry.Discovery
}

// NewConsulRegistrar creates both Registrar and Discovery.
//
// Note:
// - Registrar will auto register/deregister via kratos.
// - Discovery is used by clients via "discovery:///" endpoints.
func NewConsulParts(c *templateconf.Registry) (*ConsulParts, error) {
	if c == nil {
		return nil, nil
	}
	if !c.GetEnabled() {
		return nil, nil
	}
	if c.Type != "consul" {
		return nil, fmt.Errorf("unsupported registry type: %s", c.Type)
	}
	if c.Consul == nil {
		return nil, fmt.Errorf("registry.consul is required")
	}
	// strict YAML：要求 configs/registry.yaml 填齐字段；token 可为空。
	r, d, err := infraConsul.NewConsulRegistrar(infraConsul.Config{
		Address:   c.Consul.Address,
		Scheme:    c.Consul.Scheme,
		Token:     c.Consul.Token,
		WaitEvery: c.Consul.WaitEvery.AsDuration(),
	})
	if err != nil {
		return nil, err
	}
	return &ConsulParts{
		Registrar: r,
		Discovery: d,
	}, nil
}

func NewRegistrarFromParts(p *ConsulParts) registry.Registrar {
	if p == nil {
		return nil
	}
	return p.Registrar
}

func NewDiscoveryFromParts(p *ConsulParts) registry.Discovery {
	if p == nil {
		return nil
	}
	return p.Discovery
}
