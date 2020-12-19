package docker

import (
	"time"
)

// TimeoutConfig drives the various timeouts in the Docker backend.
type TimeoutConfig struct {
	// ContainerStart is the maximum time starting a container may take.
	ContainerStart time.Duration `json:"containerStart" yaml:"containerStart" default:"60s"`
	// ContainerStop is the maximum time to wait for a container to stop. This should always be set higher than the Docker StopTimeout.
	ContainerStop time.Duration `json:"containerStop" yaml:"containerStop" default:"60s"`
	// CommandStart sets the maximum time starting a command may take.
	CommandStart time.Duration `json:"commandStart" yaml:"commandStart" default:"60s"`
	// HTTP
	HTTP time.Duration `json:"http" yaml:"http" default:"15s"`
}
