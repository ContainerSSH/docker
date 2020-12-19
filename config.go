package docker

import (
	"fmt"
)

// Config is the base configuration structure of the DockerRun backend.
type Config struct {
	// Connection configures how to connect to dockerd
	Connection ConnectionConfig `json:"connection" yaml:"connection"`
	// Execution drives how the container and the workload is executed
	Execution ExecutionConfig `json:"execution" yaml:"execution"`
	// Timeouts configures the various timeouts when interacting with dockerd.
	Timeouts TimeoutConfig `json:"timeouts" yaml:"timeouts"`
}

// Validate validates the provided configuration and returns an error if invalid.
func (c Config) Validate() error {
	if err := c.Connection.Validate(); err != nil {
		return fmt.Errorf("invalid connection configuration (%w)", err)
	}
	if err := c.Execution.Validate(); err != nil {
		return fmt.Errorf("invalid execution configuration (%w)", err)
	}
	return nil
}
