package function

import (
	"time"

	"github.com/baetyl/baetyl-go/v2/http"
)

// Config
type Config struct {
	Server ServerConfig `yaml:"server" json:"server"`
	Client ClientConfig `yaml:"client" json:"client"`
}

// ServerConfig http server config
type ServerConfig struct {
	Host              Host `yaml:"host" json:"host"`
	http.ServerConfig `yaml:"server" json:"server"`
}

type Host struct {
	Function string `yaml:"function" json:"function" default:"baetyl-function.baetyl-edge-system"`
	Service  string `yaml:"service" json:"service" default:"baetyl-proxy.baetyl-edge-system"`
}

type ClientConfig struct {
	Http HttpConfig `yaml:"http" json:"http"`
	Grpc GrpcConfig `yaml:"grpc" json:"grpc"`
}

type HttpConfig struct {
	MaxConnsPerHost           int           `yaml:"maxConnsPerHost" json:"maxConnsPerHost" default:"512"`
	ReadTimeout               time.Duration `yaml:"readTimeout" json:"readTimeout" default:"5m"`
	MaxConnDuration           time.Duration `yaml:"maxConnDuration" json:"maxConnDuration" default:"5m"`
	MaxIdemponentCallAttempts int           `yaml:"maxIdemponentCallAttempts" json:"maxIdemponentCallAttempts" default:"3"`
}

type GrpcConfig struct {
	Port    int           `yaml:"port" json:"port" default:"80"`
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"5m"`
}
