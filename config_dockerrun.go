package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Validate validates the docker run config
//goland:noinspection GoDeprecation
func (config DockerRunConfig) Validate() error {
	if config.Host == "" {
		return fmt.Errorf("empty Docker host provided")
	}
	if err := config.Config.Validate(); err != nil {
		return fmt.Errorf("invalid dockerrun config (%w)", err)
	}
	return nil
}

//Deprecated: Switch to the more generic "docker" backend.
//goland:noinspection GoNameStartsWithPackageName,GoDeprecation
type DockerRunContainerConfig struct {
	LaunchConfig   `json:",inline" yaml:",inline"`
	Subsystems     map[string]string `json:"subsystems" yaml:"subsystems" comment:"Subsystem names and binaries map." default:"{\"sftp\":\"/usr/lib/openssh/sftp-server\"}"`
	DisableCommand bool              `json:"disableCommand" yaml:"disableCommand" comment:"Disable command execution passed from SSH"`
	Timeout        time.Duration     `json:"timeout" yaml:"timeout" comment:"Timeout for pod creation" default:"60s"`
}

//goland:noinspection GoDeprecation
func (d *DockerRunContainerConfig) UnmarshalJSON(b []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(b))
	tmp := &tmpDockerRunContainerConfig{}
	if err := decoder.Decode(tmp); err != nil {
		return err
	}

	d.DisableCommand = tmp.DisableCommand
	if err := parseRawDuration(tmp.Timeout, &d.Timeout); err != nil {
		return err
	}
	d.Subsystems = tmp.Subsystems

	lc := LaunchConfig{}
	decoder = json.NewDecoder(bytes.NewReader(b))
	if err := decoder.Decode(&lc); err != nil {
		return err
	}
	d.LaunchConfig = lc
	return nil
}

//goland:noinspection GoDeprecation
func (d *DockerRunContainerConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {

	raw := &map[string]interface{}{}
	if err := unmarshal(raw); err != nil {
		return err
	}
	substructure, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}
	tmp := &tmpDockerRunContainerConfig{}
	if err = yaml.Unmarshal(substructure, tmp); err != nil {
		return err
	}

	d.DisableCommand = tmp.DisableCommand
	if err := parseRawDuration(tmp.Timeout, &d.Timeout); err != nil {
		return err
	}
	d.Subsystems = tmp.Subsystems

	lc := LaunchConfig{}
	if err := unmarshal(&lc); err != nil {
		return err
	}
	d.LaunchConfig = lc
	return nil
}

// Validate validates the container config
//goland:noinspection GoDeprecation
func (d *DockerRunContainerConfig) Validate() error {
	return d.LaunchConfig.Validate()
}

type tmpDockerRunContainerConfig struct {
	Subsystems     map[string]string `json:"subsystems" yaml:"subsystems" comment:"Subsystem names and binaries map." default:"{\"sftp\":\"/usr/lib/openssh/sftp-server\"}"`
	DisableCommand bool              `json:"disableCommand" yaml:"disableCommand" comment:"Disable command execution passed from SSH"`
	Timeout        interface{}       `json:"timeout" yaml:"timeout" comment:"Timeout for pod creation" default:"60s"`
}
