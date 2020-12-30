package docker

import (
	"net"

	"github.com/containerssh/log"
	"github.com/containerssh/metrics"
	"github.com/containerssh/sshserver"
)

// NewDockerRun creates a new NetworkConnectionHandler based on the deprecated "dockerrun" config structure.
// Deprecated: use New instead
//goland:noinspection GoDeprecation
func NewDockerRun(
	client net.TCPAddr,
	connectionID string,
	legacyConfig DockerRunConfig,
	logger log.Logger,
	backendRequestsMetric metrics.SimpleCounter,
	backendFailuresMetric metrics.SimpleCounter,
) (sshserver.NetworkConnectionHandler, error) {
	logger.Warningf(
		"You are using the dockerrun backend deprecated since ContainerSSH 0.4. This backend will be removed " +
			"in the future. Please switch to the new docker backend as soon as possible. " +
			"See https://containerssh.io/deprecations/dockerrun for details.",
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
		DisableAgent:    true,
		ImagePullPolicy: ImagePullPolicyAlways,
		disableCommand:  legacyConfig.Config.DisableCommand,
	}
	config.Timeouts = TimeoutConfig{
		ContainerStart: legacyConfig.Config.Timeout,
		ContainerStop:  legacyConfig.Config.Timeout,
		CommandStart:   legacyConfig.Config.Timeout,
	}

	return New(
		client,
		connectionID,
		config,
		logger,
		backendRequestsMetric,
		backendFailuresMetric,
	)
}
