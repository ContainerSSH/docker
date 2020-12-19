// +build windows

package docker

import (
	"fmt"
)

// ConnectionConfig configures how to connect to dockerd.
type ConnectionConfig struct {
	// Host is the docker connect URL
	Host string `json:"host" yaml:"host" default:"npipe:////./pipe/docker_engine"`
	// CaCert is the CA certificate for Docker connection embedded in the configuration in PEM format.
	CaCert string `json:"cacert" yaml:"cacert"`
	// Cert is the client certificate in PEM format embedded in the configuration.
	Cert string `json:"cert" yaml:"cert"`
	// Key is the client key in PEM format embedded in the configuration.
	Key string `json:"key" yaml:"key"`
}

func (c ConnectionConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("missing host")
	}
	return nil
}
