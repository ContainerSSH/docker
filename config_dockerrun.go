package docker

import (
	"time"
)

//goland:noinspection GoNameStartsWithPackageName
//Deprecated: Switch to the more generic "docker" backend.
//goland:noinspection GoNameStartsWithPackageName,GoDeprecation
type DockerRunContainerConfig struct {
	LaunchConfig   `json:",inline" yaml:",inline"`
	Subsystems     map[string]string `json:"subsystems" yaml:"subsystems" comment:"Subsystem names and binaries map." default:"{\"sftp\":\"/usr/lib/openssh/sftp-server\"}"`
	DisableCommand bool              `json:"disableCommand" yaml:"disableCommand" comment:"Disable command execution passed from SSH"`
	Timeout        time.Duration     `json:"timeout" yaml:"timeout" comment:"Timeout for pod creation" default:"60s"`
}
