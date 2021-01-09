package docker

import (
	"net"
	"sync"

	"github.com/containerssh/log"
	"github.com/containerssh/metrics"
	"github.com/containerssh/sshserver"
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
		logger.Warningf("ContainerSSH Guest Agent support is disabled. Some functions will not work.")
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
