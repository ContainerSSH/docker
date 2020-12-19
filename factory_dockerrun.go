package docker

import (
	"net"

	"github.com/containerssh/log"
	"github.com/containerssh/sshserver"
)

// NewDockerRun creates a new NetworkConnectionHandler based on the deprecated "dockerrun" config structure.
//goland:noinspection GoDeprecation
func NewDockerRun(
	client net.TCPAddr,
	connectionID string,
	legacyConfig DockerRunConfig,
	logger log.Logger,
) (sshserver.NetworkConnectionHandler, error) {
	logger.Warningf(
		"You are using the deprecated \"dockerrun\" backend which will be removed in future ContainerSSH " +
			"versions. Please switch to the new \"docker\" backend. Please read the deprecation notice at " +
			"https://containerssh.io/deprecations/dockerrun for instructions on upgrading.",
	)

	config := Config{}

	config.Connection = ConnectionConfig{
		legacyConfig.Host,
		legacyConfig.CaCert,
		legacyConfig.Cert,
		legacyConfig.Key,
	}
	config.Execution = ExecutionConfig{
		Launch: LaunchConfig{
			ContainerConfig: legacyConfig.Config.ContainerConfig,
			HostConfig:      legacyConfig.Config.HostConfig,
			NetworkConfig:   legacyConfig.Config.NetworkConfig,
			Platform:        legacyConfig.Config.Platform,
			ContainerName:   legacyConfig.Config.ContainerName,
		},
		Mode:            ExecutionModeSession,
		Subsystems:      legacyConfig.Config.Subsystems,
		ShellCommand:    nil,
		IdleCommand:     nil,
		ImagePullPolicy: ImagePullPolicyAlways,
		disableCommand:  legacyConfig.Config.DisableCommand,
	}
	config.Timeouts = TimeoutConfig{
		ContainerStart: legacyConfig.Config.Timeout,
		ContainerStop:  legacyConfig.Config.Timeout,
		CommandStart:   legacyConfig.Config.Timeout,
	}

	return New(client, connectionID, config, logger)
}
