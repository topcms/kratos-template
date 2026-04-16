package conf

import "time"

// Bootstrap 对应 configs/config.yaml 根结构。
type Bootstrap struct {
	Server *Server `json:"server" yaml:"server"`
	Data   *Data   `json:"data" yaml:"data"`
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
}

// Database 数据库。
type Database struct {
	Driver string `json:"driver" yaml:"driver"`
	Source string `json:"source" yaml:"source"`
}

// Redis 缓存。
type Redis struct {
	Network      string        `json:"network" yaml:"network"`
	Addr         string        `json:"addr" yaml:"addr"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
}
