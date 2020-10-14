package function

import (
	"time"

	"github.com/baetyl/baetyl-go/v2/http"
)

// Config
type Config struct {
	Server http.ServerConfig `yaml:"server" json:"server"`
	Client ClientConfig      `yaml:"client" json:"client"`
}

type ClientConfig struct {
	Grpc GrpcConfig `yaml:"grpc" json:"grpc"`
}

type GrpcConfig struct {
	Port    int           `yaml:"port" json:"port" default:"80"`
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"5m"`
	Retries int           `yaml:"retries" json:"retries" default:"3"`
}
