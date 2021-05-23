package docker_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/containerssh/structutils"
	"github.com/docker/docker/api/types/container"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/containerssh/docker"
)

// TestYAMLSerialization tests if the configuration structure can be serialized and then deserialized to/from YAML.
func TestYAMLSerialization(t *testing.T) {
	t.Parallel()

	// region Setup
	config := &docker.Config{}
	newCfg := &docker.Config{}
	structutils.Defaults(config)

	buf := &bytes.Buffer{}
	// endregion

	// region Save
	yamlEncoder := yaml.NewEncoder(buf)
	assert.NoError(t, yamlEncoder.Encode(config))
	// endregion

	// region Load
	yamlDecoder := yaml.NewDecoder(buf)
	yamlDecoder.KnownFields(true)
	assert.NoError(t, yamlDecoder.Decode(newCfg))
	// endregion

	// region Assert

	diff := cmp.Diff(
		config,
		newCfg,
		cmp.AllowUnexported(docker.ExecutionConfig{}),
		cmpopts.EquateEmpty(),
	)
	assert.Empty(t, diff)
	// endregion
}

// TestJSONSerialization tests if the configuration structure can be serialized and then deserialized to/from JSON.
func TestJSONSerialization(t *testing.T) {
	t.Parallel()

	// region Setup
	config := &docker.Config{}
	newCfg := &docker.Config{}
	structutils.Defaults(config)

	buf := &bytes.Buffer{}
	// endregion

	// region Save
	jsonEncoder := json.NewEncoder(buf)
	assert.NoError(t, jsonEncoder.Encode(config))
	// endregion

	// region Load
	jsonDecoder := json.NewDecoder(buf)
	jsonDecoder.DisallowUnknownFields()
	assert.NoError(t, jsonDecoder.Decode(newCfg))
	// endregion

	// region Assert

	diff := cmp.Diff(
		config,
		newCfg,
		cmp.AllowUnexported(docker.ExecutionConfig{}),
		cmpopts.EquateEmpty(),
	)
	assert.Empty(t, diff)
	// endregion
}

// TestLaunchInline tests if the launch config is properly inlined when marshaling.
func TestLaunchInlineMarshal(t *testing.T) {
	config := &docker.Config{}
	config.Execution.Launch.ContainerConfig = &container.Config{}
	config.Execution.Launch.ContainerConfig.Image = "containerssh/test"

	marshalledConfig, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	unmarshalledConfig := map[string]interface{}{}
	if err := json.Unmarshal(marshalledConfig, &unmarshalledConfig); err != nil {
		t.Fatal(err)
	}

	execution := unmarshalledConfig["execution"].(map[string]interface{})
	cnt := execution["container"].(map[string]interface{})

	if cnt["Image"].(string) != "containerssh/test" {
		t.Fatal("image is not set in output")
	}
}

// TestLaunchInline tests if the launch config is properly inlined when unmarshaling
func TestLaunchInlineUnmarshal(t *testing.T) {
	data := `{
		"execution": {
			"container": {
				"image": "containerssh/test"
			}
		}
	}`
	config := &docker.Config{}
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		t.Fatal(err)
	}

	if config.Execution.Launch.ContainerConfig.Image != "containerssh/test" {
		t.Fatal("image is not set in output")
	}
}
