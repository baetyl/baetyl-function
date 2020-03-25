package main

import (
	"github.com/baetyl/baetyl-go/utils"
	"time"
)

// Config
type Config struct {
	Server ServerConfig `yaml:"server" json:"server"`
	Client ClientConfig `yaml:"client" json:"client"`
}

// ServerConfig http server config
type ServerConfig struct {
	Address           string `yaml:"address" json:"address" default:":8080"`
	utils.Certificate `yaml:",inline" json:",inline"`
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
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"5m"`
}
