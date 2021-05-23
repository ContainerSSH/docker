// +build linux freebsd openbsd darwin

package docker

//DockerRunConfig describes the old ContainerSSH 0.3 configuration format that can still be read and used.
//Deprecated: Switch to the more generic "docker" backend.
//goland:noinspection GoNameStartsWithPackageName,GoDeprecation
type DockerRunConfig struct {
	Host   string                   `json:"host,omitempty" yaml:"host" comment:"Docker connect URL" default:"unix:///var/run/docker.sock"`
	CaCert string                   `json:"cacert,omitempty" yaml:"cacert" comment:"CA certificate for Docker connection embedded in the configuration in PEM format."`
	Cert   string                   `json:"cert,omitempty" yaml:"cert" comment:"Client certificate in PEM format embedded in the configuration."`
	Key    string                   `json:"key,omitempty" yaml:"key" comment:"Client key in PEM format embedded in the configuration."`
	Config DockerRunContainerConfig `json:"config,omitempty" yaml:"config" comment:"Config configuration"`
}
