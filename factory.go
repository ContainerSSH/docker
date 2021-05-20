package docker

import (
	"net"
	"reflect"
	"sync"

	"github.com/containerssh/log"
	"github.com/containerssh/metrics"
	"github.com/containerssh/sshserver/v2"
	"github.com/containerssh/structutils"
)

// New creates a new NetworkConnectionHandler for a specific client.
func New(
	client net.TCPAddr,
	connectionID string,
	config Config,
	logger log.Logger,
	backendRequestsMetric metrics.SimpleCounter,
	backendFailuresMetric metrics.SimpleCounter,
) (
	sshserver.NetworkConnectionHandler,
	error,
) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if config.Execution.DisableAgent {
		logger.Warning(log.NewMessage(EGuestAgentDisabled, "ContainerSSH Guest Agent support is disabled. Some functions will not work."))
		defaultCfg := &Config{}
		structutils.Defaults(defaultCfg)
		if config.Execution.Mode == ExecutionModeConnection && reflect.DeepEqual(config.Execution.IdleCommand, defaultCfg.Execution.IdleCommand) {
			logger.Warning(log.NewMessage(EGuestAgentDisabled, "ContainerSSH Guest Agent support is disabled, but the execution mode is set to connection and the idle command still points to the guest agent to provide an init program. This is very likely to break since you most likely don't have the guest agent installed."))
		}
	}

	return &networkHandler{
		mutex:        &sync.Mutex{},
		client:       client,
		connectionID: connectionID,
		config:       config,
		logger:       logger,
		disconnected: false,
		dockerClientFactory: &dockerV20ClientFactory{
			backendFailuresMetric: backendFailuresMetric,
			backendRequestsMetric: backendRequestsMetric,
		},
		done: make(chan struct{}),
	}, nil
}
