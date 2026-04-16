package server

import (
	"github.com/topcms/kratos-template/internal/conf"

	infrajwt "github.com/topcms/kratos-infra/auth/jwt"
	infraauth "github.com/topcms/kratos-infra/middleware/auth"
)

// NewTokenValidator adapts template config to infra jwt validator.
func NewTokenValidator(c *conf.Auth) (infraauth.TokenValidator, error) {
	if c == nil || c.JWT == nil {
		return nil, nil
	}
	return infrajwt.NewTokenValidator(infrajwt.Config{
		Enabled:       c.JWT.Enabled,
		SigningMethod: c.JWT.SigningMethod,
		Secret:        c.JWT.Secret,
		Issuer:        c.JWT.Issuer,
		Audience:      c.JWT.Audience,
	})
}
