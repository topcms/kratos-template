package server

import (
	"github.com/topcms/kratos-template/internal/conf"

	infrajwt "github.com/topcms/kratos-infra/auth/jwt"
	infraauth "github.com/topcms/kratos-infra/middleware/auth"
)

// NewTokenValidator adapts template config to infra jwt validator.
func NewTokenValidator(c *conf.Auth) (infraauth.TokenValidator, error) {
	if c == nil || c.Jwt == nil {
		return nil, nil
	}
	return infrajwt.NewTokenValidator(infrajwt.Config{
		Enabled:       c.Jwt.Enabled,
		SigningMethod: c.Jwt.SigningMethod,
		Secret:        c.Jwt.Secret,
		Issuer:        c.Jwt.Issuer,
		Audience:      c.Jwt.Audience,
	})
}
