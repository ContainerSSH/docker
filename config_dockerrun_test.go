package docker_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/containerssh/docker/v2"
)

// TestUnmarshalYAML03 tests the ContainerSSH 0.3 compatibility. It checks if a YAML fragment from 0.3 can still be
// unmarshalled.
func TestUnmarshalYAML03(t *testing.T) {
	t.Parallel()

	testFile, err := os.Open("testdata/config-0.3.yaml")
	assert.NoError(t, err)
	unmarshaller := yaml.NewDecoder(testFile)
	unmarshaller.KnownFields(true)
	//goland:noinspection GoDeprecation
	config := docker.DockerRunConfig{}
	assert.NoError(t, unmarshaller.Decode(&config))
	assert.Equal(t, false, config.Config.DisableCommand)
	assert.Equal(t, "/usr/lib/openssh/sftp-server", config.Config.Subsystems["sftp"])
	assert.Equal(t, "containerssh/containerssh-guest-image", config.Config.LaunchConfig.ContainerConfig.Image)
	assert.Equal(t, 60*time.Second, config.Config.Timeout)
}

// TestUnmarshalYAML03 tests the ContainerSSH 0.3 compatibility. It checks if a JSON fragment from 0.3 can still be
// unmarshalled.
func TestUnmarshalJSON03(t *testing.T) {
	t.Parallel()

	testFile, err := os.Open("testdata/config-0.3.json")
	assert.NoError(t, err)
	unmarshaller := json.NewDecoder(testFile)
	unmarshaller.DisallowUnknownFields()
	//goland:noinspection GoDeprecation
	config := docker.DockerRunConfig{}
	assert.NoError(t, unmarshaller.Decode(&config))
	assert.Equal(t, false, config.Config.DisableCommand)
	assert.Equal(t, "/usr/lib/openssh/sftp-server", config.Config.Subsystems["sftp"])
	assert.Equal(t, "containerssh/containerssh-guest-image", config.Config.LaunchConfig.ContainerConfig.Image)
	assert.Equal(t, 60*time.Second, config.Config.Timeout)
}
