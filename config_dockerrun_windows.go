// +build windows

package docker

//DockerRunConfig describes the old ContainerSSH 0.3 configuration format that can still be read and used.
//Deprecated: Switch to the more generic "docker" backend.
//goland:noinspection GoNameStartsWithPackageName,GoDeprecation
type DockerRunConfig struct {
	Host   string                   `json:"host" yaml:"host" comment:"Docker connect URL" default:"npipe:////./pipe/docker_engine"`
	CaCert string                   `json:"cacert" yaml:"cacert" comment:"CA certificate for Docker connection embedded in the configuration in PEM format."`
	Cert   string                   `json:"cert" yaml:"cert" comment:"Client certificate in PEM format embedded in the configuration."`
	Key    string                   `json:"key" yaml:"key" comment:"Client key in PEM format embedded in the configuration."`
	Config DockerRunContainerConfig `json:"config" yaml:"config" comment:"Config configuration"`
}
