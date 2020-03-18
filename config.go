package main

import (
	"github.com/baetyl/baetyl-go/utils"
	"time"
)

// ServerInfo http server config
type ServerConfig struct {
	Address           string        `yaml:"address" json:"address" default:":8080"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout" default:"5m"`
	utils.Certificate `yaml:",inline" json:",inline"`
}
