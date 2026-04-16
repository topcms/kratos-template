package conf

import "time"

// Bootstrap 对应 configs/config.yaml 根结构。
type Bootstrap struct {
	Server   *Server   `json:"server" yaml:"server"`
	Data     *Data     `json:"data" yaml:"data"`
	Auth     *Auth     `json:"auth" yaml:"auth"`
	Registry *Registry `json:"registry" yaml:"registry"`
}

// Server 监听与超时。
type Server struct {
	Http *HTTP `json:"http" yaml:"http"`
	Grpc *GRPC `json:"grpc" yaml:"grpc"`
}

// HTTP 配置。
type HTTP struct {
	Network string        `json:"network" yaml:"network"`
	Addr    string        `json:"addr" yaml:"addr"`
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
}

// GRPC 配置。
type GRPC struct {
	Network string        `json:"network" yaml:"network"`
	Addr    string        `json:"addr" yaml:"addr"`
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
}

// Data 数据层（库、缓存等）。
type Data struct {
	Database *Database `json:"database" yaml:"database"`
	Redis    *Redis    `json:"redis" yaml:"redis"`
	Remote   *Remote   `json:"remote" yaml:"remote"`
}

// Database 数据库。
type Database struct {
	Driver string `json:"driver" yaml:"driver"`
	// Source 用作 MySQL DSN（透传给 gorm/mysql）
	Source string `json:"source" yaml:"source"`
	// 连接池参数：用于严格对齐 kratos-infra/mysql.NewDB()
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
}

// Redis 缓存。
type Redis struct {
	Addr         string        `json:"addr" yaml:"addr"`
	Password     string        `json:"password" yaml:"password"`
	DB           int           `json:"db" yaml:"db"`
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
}

// Remote 服务端调用配置（用于演示 discovery:///xxx 的 Dial）。
type Remote struct {
	UserService *RemoteUserService `json:"user_service" yaml:"user_service"`
}

type RemoteUserService struct {
	// ServiceName 必须与被调用服务的 kratos.Name 一致（即 registry 的 ServiceInstance.Name）。
	ServiceName string        `json:"service_name" yaml:"service_name"`
	DialTimeout time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
}

// Auth 鉴权配置。
type Auth struct {
	JWT *JWT `json:"jwt" yaml:"jwt"`
}

// JWT 配置（当前通过 kratos-infra/auth/jwt 支持 HS256）。
type JWT struct {
	Enabled       bool     `json:"enabled" yaml:"enabled"`
	SigningMethod string   `json:"signing_method" yaml:"signing_method"`
	Secret        string   `json:"secret" yaml:"secret"`
	Issuer        string   `json:"issuer" yaml:"issuer"`
	Audience      []string `json:"audience" yaml:"audience"`
}

// Registry 服务注册/发现配置（对应 configs/registry.yaml）。
type Registry struct {
	Type   string  `json:"type" yaml:"type"`
	Consul *Consul `json:"consul" yaml:"consul"`
}

type Consul struct {
	Address   string        `json:"address" yaml:"address"`
	Scheme    string        `json:"scheme" yaml:"scheme"`
	Token     string        `json:"token" yaml:"token"`
	WaitEvery time.Duration `json:"wait_every" yaml:"wait_every"`
}
