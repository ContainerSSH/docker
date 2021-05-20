package docker

import (
	"encoding/json"
	"fmt"
)

// ExecutionMode determines when a container is launched.
// ExecutionModeConnection launches one container per SSH connection (default), while ExecutionModeSession launches
// one container per SSH session.
type ExecutionMode string

const (
	// ExecutionModeConnection launches one container per SSH connection.
	ExecutionModeConnection ExecutionMode = "connection"
	// ExecutionModeSession launches one container per SSH session (multiple containers per connection).
	ExecutionModeSession ExecutionMode = "session"
)

// Validate validates the execution config.
func (e ExecutionMode) Validate() error {
	switch e {
	case ExecutionModeConnection:
		fallthrough
	case ExecutionModeSession:
		return nil
	default:
		return fmt.Errorf("invalid execution mode: %s", e)
	}
}

// ExecutionConfig contains the configuration of what container to run in Docker.
//goland:noinspection GoVetStructTag
type ExecutionConfig struct {
	// Launch contains the Docker-specific launch configuration.
	Launch LaunchConfig `json:",inline,omitempty" yaml:",inline"`
	// Mode influences how commands are executed.
	//
	// - If ExecutionModeConnection is chosen (default) a new container is launched per connection. In this mode
	//   sessions are executed using the "docker exec" functionality and the main container console runs a script that
	//   waits for a termination signal.
	// - If ExecutionModeSession is chosen a new container is launched per session, leading to potentially multiple
	//   containers per connection. In this mode the program is launched directly as the main process of the container.
	//   When configuring this mode you should explicitly configure the "cmd" option to an empty list if you want the
	//   default command in the container to launch.
	Mode ExecutionMode `json:"mode,omitempty" yaml:"mode" default:"connection"`

	// IdleCommand is the command that runs as the first process in the container in ExecutionModeConnection. Ignored in ExecutionModeSession.
	IdleCommand []string `json:"idleCommand,omitempty" yaml:"idleCommand" comment:"Run this command to wait for container exit" default:"[\"/usr/bin/containerssh-agent\", \"wait-signal\", \"--signal\", \"INT\", \"--signal\", \"TERM\"]"`
	// ShellCommand is the command used for launching shells when the container is in ExecutionModeConnection. Ignored in ExecutionModeSession.
	ShellCommand []string `json:"shellCommand,omitempty" yaml:"shellCommand" comment:"Run this command as a default shell." default:"[\"/bin/bash\"]"`
	// AgentPath contains the path to the ContainerSSH Guest Agent.
	AgentPath string `json:"agentPath,omitempty" yaml:"agentPath" default:"/usr/bin/containerssh-agent"`
	// DisableAgent enables using the ContainerSSH Guest Agent.
	DisableAgent bool `json:"disableAgent,omitempty" yaml:"disableAgent"`
	// Subsystems contains a map of subsystem names and their corresponding binaries in the container.
	Subsystems map[string]string `json:"subsystems" yaml:"subsystems" comment:"Subsystem names and binaries map." default:"{\"sftp\":\"/usr/lib/openssh/sftp-server\"}"`

	// ImagePullPolicy controls when to pull container images.
	ImagePullPolicy ImagePullPolicy `json:"imagePullPolicy" yaml:"imagePullPolicy" comment:"Image pull policy" default:"IfNotPresent"`

	// disableCommand is a configuration option to support legacy command disabling from the dockerrun config.
	// See https://containerssh.io/deprecations/dockerrun for details.
	disableCommand bool `json:"-" yaml:"-"`
}

type tmpExecutionConfig struct {
	Mode ExecutionMode `json:"mode,omitempty" yaml:"mode" default:"connection"`
	IdleCommand []string `json:"idleCommand,omitempty" yaml:"idleCommand" comment:"Run this command to wait for container exit" default:"[\"/usr/bin/containerssh-agent\", \"wait-signal\", \"--signal\", \"INT\", \"--signal\", \"TERM\"]"`
	ShellCommand []string `json:"shellCommand,omitempty" yaml:"shellCommand" comment:"Run this command as a default shell." default:"[\"/bin/bash\"]"`
	AgentPath string `json:"agentPath,omitempty" yaml:"agentPath" default:"/usr/bin/containerssh-agent"`
	DisableAgent bool `json:"disableAgent,omitempty" yaml:"disableAgent"`
	Subsystems map[string]string `json:"subsystems" yaml:"subsystems" comment:"Subsystem names and binaries map." default:"{\"sftp\":\"/usr/lib/openssh/sftp-server\"}"`
	ImagePullPolicy ImagePullPolicy `json:"imagePullPolicy" yaml:"imagePullPolicy" comment:"Image pull policy" default:"IfNotPresent"`
}

// UnmarshalJSON provides inlining capabilities for LaunchConfig
func (c *ExecutionConfig) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &c.Launch); err != nil {
		return fmt.Errorf("failed to decode launch config (%w)", err)
	}

	cfg := tmpExecutionConfig{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	c.Mode = cfg.Mode
	c.IdleCommand = cfg.IdleCommand
	c.ShellCommand = cfg.ShellCommand
	c.AgentPath = cfg.AgentPath
	c.DisableAgent = cfg.DisableAgent
	c.Subsystems = cfg.Subsystems
	c.ImagePullPolicy = cfg.ImagePullPolicy
	return nil
}

func (c ExecutionConfig) MarshalJSON() ([]byte, error) {
	launchData, err := json.Marshal(c.Launch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal launch config (%w)", err)
	}
	unmarshalledLaunchData := map[string]interface{}{}
	if err := json.Unmarshal(launchData, &unmarshalledLaunchData); err != nil {
		return nil, err
	}
	cfg := tmpExecutionConfig{
		Mode:            c.Mode,
		IdleCommand:     c.IdleCommand,
		ShellCommand:    c.ShellCommand,
		AgentPath:       c.AgentPath,
		DisableAgent:    c.DisableAgent,
		Subsystems:      c.Subsystems,
		ImagePullPolicy: c.ImagePullPolicy,
	}
	cfgData, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	unmarshalledCfgData := map[string]interface{}{}
	if err := json.Unmarshal(cfgData, &unmarshalledCfgData); err != nil {
		return nil, err
	}
	for k,v := range unmarshalledLaunchData {
		unmarshalledCfgData[k] = v
	}
	return json.Marshal(unmarshalledCfgData)
}

// Validate validates the docker config structure.
func (c ExecutionConfig) Validate() error {
	if c.Mode == ExecutionModeConnection && len(c.IdleCommand) == 0 {
		return fmt.Errorf("idle command required for execution mode \"connection\"")
	}
	if c.Mode == ExecutionModeConnection && len(c.ShellCommand) == 0 {
		return fmt.Errorf("shell command required for execution mode \"connection\"")
	}
	switch c.Mode {
	case ExecutionModeSession:
		if c.Launch.HostConfig != nil && !c.Launch.HostConfig.RestartPolicy.IsNone() {
			return fmt.Errorf(
				"unsupported restart policy for execution mode \"session\": %s (session containers may not restart)",
				c.Launch.HostConfig.RestartPolicy.Name,
			)
		}
	}
	if err := c.ImagePullPolicy.Validate(); err != nil {
		return err
	}
	if err := c.Launch.Validate(); err != nil {
		return err
	}
	if err := c.Mode.Validate(); err != nil {
		return err
	}
	return nil
}

// ImagePullPolicy drives how and when images are pulled. The values are closely aligned with the Kubernetes image pull
// policy.
//
// - ImagePullPolicyAlways means that the container image will be pulled on every connection.
// - ImagePullPolicyIfNotPresent means the image will be pulled if the image is not present locally, an empty tag, or
//	 the "latest" tag was specified.
// - ImagePullPolicyNever means that the image will be never pulled, and if the image is not available locally the
//	 connection will fail.
type ImagePullPolicy string

const (
	// ImagePullPolicyAlways means that the container image will be pulled on every connection.
	ImagePullPolicyAlways ImagePullPolicy = "Always"
	// ImagePullPolicyIfNotPresent means the image will be pulled if the image is not present locally, an empty tag, or
	// the "latest" tag was specified.
	ImagePullPolicyIfNotPresent ImagePullPolicy = "IfNotPresent"
	// ImagePullPolicyNever means that the image will be never pulled, and if the image is not available locally the
	// connection will fail.
	ImagePullPolicyNever ImagePullPolicy = "Never"
)

// Validate checks if the given image pull policy is valid.
func (p ImagePullPolicy) Validate() error {
	switch p {
	case ImagePullPolicyAlways:
		fallthrough
	case ImagePullPolicyIfNotPresent:
		fallthrough
	case ImagePullPolicyNever:
		return nil
	default:
		return fmt.Errorf("invalid image pull policy: %s", p)
	}
}
