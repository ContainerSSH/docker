// +build windows

package docker

//DockerRunConfig describes the old ContainerSSH 0.3 configuration format that can still be read and used.
//Deprecated: Switch to the more generic "docker" backend.
//goland:noinspection GoNameStartsWithPackageName,GoDeprecation
type DockerRunConfig struct {
	Host   string                   `json:"host,omitempty" yaml:"host,omitempty" comment:"Docker connect URL" default:"npipe:////./pipe/docker_engine"`
	CaCert string                   `json:"cacert,omitempty" yaml:"cacert,omitempty" comment:"CA certificate for Docker connection embedded in the configuration in PEM format."`
	Cert   string                   `json:"cert,omitempty" yaml:"cert,omitempty" comment:"Client certificate in PEM format embedded in the configuration."`
	Key    string                   `json:"key,omitempty" yaml:"key,omitempty" comment:"Client key in PEM format embedded in the configuration."`
	Config DockerRunContainerConfig `json:"config,omitempty" yaml:"config,omitempty" comment:"Config configuration"`
}
