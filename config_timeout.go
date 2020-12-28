package docker

import (
	"bytes"
	"encoding/json"
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
	// Signal sets the maximum time sending a signal may take.
	Signal time.Duration `json:"signal" yaml:"signal" default:"60s"`
	// Signal sets the maximum time setting the window size may take.
	Window time.Duration `json:"window" yaml:"window" default:"60s"`
	// HTTP
	HTTP time.Duration `json:"http" yaml:"http" default:"15s"`
}

type tmpTimeoutConfig struct {
	// ContainerStart is the maximum time starting a container may take.
	ContainerStart interface{} `json:"containerStart" yaml:"containerStart" default:"60s"`
	// ContainerStop is the maximum time to wait for a container to stop. This should always be set higher than the Docker StopTimeout.
	ContainerStop interface{} `json:"containerStop" yaml:"containerStop" default:"60s"`
	// CommandStart sets the maximum time starting a command may take.
	CommandStart interface{} `json:"commandStart" yaml:"commandStart" default:"60s"`
	// Signal sets the maximum time sending a signal may take.
	Signal interface{} `json:"signal" yaml:"signal" default:"60s"`
	// Signal sets the maximum time setting the window size may take.
	Window interface{} `json:"window" yaml:"window" default:"60s"`
	// HTTP
	HTTP interface{} `json:"http" yaml:"http" default:"15s"`
}

// UnmarshalJSON takes a JSON byte array and unmarshalls it into a structure.
func (t *TimeoutConfig) UnmarshalJSON(b []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(b))
	tmp := &tmpTimeoutConfig{}
	if err := decoder.Decode(tmp); err != nil {
		return err
	}

	return t.unmarshalTmp(tmp)
}

// UnmarshalYAML takes a YAML byte array and unmarshalls it into a structure.
func (t *TimeoutConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	tmp := &tmpTimeoutConfig{}
	if err := unmarshal(tmp); err != nil {
		return err
	}

	return t.unmarshalTmp(tmp)
}

func (t *TimeoutConfig) unmarshalTmp(tmp *tmpTimeoutConfig) error {
	if err := parseRawDuration(tmp.ContainerStart, &t.ContainerStart); err != nil {
		return err
	}
	if err := parseRawDuration(tmp.ContainerStop, &t.ContainerStop); err != nil {
		return err
	}
	if err := parseRawDuration(tmp.CommandStart, &t.CommandStart); err != nil {
		return err
	}
	if err := parseRawDuration(tmp.Signal, &t.Signal); err != nil {
		return err
	}
	if err := parseRawDuration(tmp.Window, &t.Window); err != nil {
		return err
	}
	if err := parseRawDuration(tmp.HTTP, &t.HTTP); err != nil {
		return err
	}
	return nil
}
