package docker

import (
	"github.com/containerssh/log"
)

// Config is the base configuration structure of the DockerRun backend.
type Config struct {
	// Connection configures how to connect to dockerd
	Connection ConnectionConfig `json:"connection,omitempty" yaml:"connection,omitempty"`
	// Execution drives how the container and the workload is executed
	Execution ExecutionConfig `json:"execution,omitempty" yaml:"execution,omitempty"`
	// Timeouts configures the various timeouts when interacting with dockerd.
	Timeouts TimeoutConfig `json:"timeouts,omitempty" yaml:"timeouts,omitempty"`
}

// Validate validates the provided configuration and returns an error if invalid.
func (c Config) Validate() error {
	if err := c.Connection.Validate(); err != nil {
		return log.Wrap(err, EConfigError, "invalid connection configuration")
	}
	if err := c.Execution.Validate(); err != nil {
		return log.Wrap(err, EConfigError, "invalid execution configuration")
	}
	return nil
}
