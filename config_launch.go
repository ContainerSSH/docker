package docker

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"gopkg.in/yaml.v3"
)

// LaunchConfig contains the container configuration for the Docker client version 20.
type LaunchConfig struct {
	// ContainerConfig contains container-specific configuration options.
	ContainerConfig *container.Config `json:"container" yaml:"container" comment:"Config configuration." default:"{\"image\":\"containerssh/containerssh-guest-image\"}"`
	// HostConfig contains the host-specific configuration options.
	HostConfig *container.HostConfig `json:"host" yaml:"host" comment:"Host configuration"`
	// NetworkConfig contains the network settings.
	NetworkConfig *network.NetworkingConfig `json:"network" yaml:"network" comment:"Network configuration"`
	// Platform contains the platform specification.
	Platform *specs.Platform `json:"platform" yaml:"platform" comment:"Platform specification"`
	// ContainerName is the name of the container to launch. It is recommended to leave this empty, otherwise
	// ContainerSSH may not be able to start the container if a container with the same name already exists.
	ContainerName string `json:"containername" yaml:"containername" comment:"Name for the container to be launched"`
}

type tmpLaunchConfig struct {
	// ContainerConfig contains container-specific configuration options.
	ContainerConfig *container.Config `json:"container" yaml:"container"`
	// HostConfig contains the host-specific configuration options.
	HostConfig *container.HostConfig `json:"host" yaml:"host"`
	// NetworkConfig contains the network settings.
	NetworkConfig *network.NetworkingConfig `json:"network" yaml:"network"`
	// Platform contains the platform specification.
	Platform *specs.Platform `json:"platform" yaml:"platform"`
	// ContainerName is the name of the container to launch. It is recommended to leave this empty, otherwise
	// ContainerSSH may not be able to start the container if a container with the same name already exists.
	ContainerName string `json:"containername" yaml:"containername"`
}

// UnmarshalJSON implements the special unmarshalling of the LaunchConfig that ignores unknown fields.
// This is needed because Docker treats removing fields as backwards-compatible.
// See https://github.com/moby/moby/pull/39158#issuecomment-489704731
func (l *LaunchConfig) UnmarshalJSON(b []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(b))
	tmp := &tmpLaunchConfig{}
	if err := decoder.Decode(tmp); err != nil {
		return err
	}
	l.ContainerConfig = tmp.ContainerConfig
	l.HostConfig = tmp.HostConfig
	l.NetworkConfig = tmp.NetworkConfig
	l.Platform = tmp.Platform
	l.ContainerName = tmp.ContainerName
	return nil
}

// UnmarshalYAML implements the special unmarshalling of the LaunchConfig that ignores unknown fields.
// This is needed because Docker treats removing fields as backwards-compatible.
// See https://github.com/moby/moby/pull/39158#issuecomment-489704731
func (l *LaunchConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	lc := &map[string]interface{}{}
	if err := unmarshal(lc); err != nil {
		return err
	}
	substructure, err := yaml.Marshal(lc)
	if err != nil {
		return err
	}
	tmp := &tmpLaunchConfig{}
	if err = yaml.Unmarshal(substructure, tmp); err != nil {
		return err
	}
	l.ContainerConfig = tmp.ContainerConfig
	l.HostConfig = tmp.HostConfig
	l.NetworkConfig = tmp.NetworkConfig
	l.Platform = tmp.Platform
	l.ContainerName = tmp.ContainerName
	return nil
}

// Validate validates the launch configuration.
func (l *LaunchConfig) Validate() error {
	if l.ContainerConfig == nil {
		return fmt.Errorf("no container config provided")
	}
	if l.ContainerConfig.Image == "" {
		return fmt.Errorf("no image name provided")
	}
	return nil
}
