// +build linux freebsd openbsd darwin

package docker

// ConnectionConfig configures how to connect to dockerd.
type ConnectionConfig struct {
	// Host is the docker connect URL.
	Host string `json:"host,omitempty" yaml:"host" default:"unix:///var/run/docker.sock"`
	// CaCert is the CA certificate for Docker connection embedded in the configuration in PEM format.
	CaCert string `json:"cacert,omitempty" yaml:"cacert"`
	// Cert is the client certificate in PEM format embedded in the configuration.
	Cert string `json:"cert,omitempty" yaml:"cert"`
	// Key is the client key in PEM format embedded in the configuration.
	Key string `json:"key,omitempty" yaml:"key"`
}
