package docker

import (
	"net"
	"sync"

	"github.com/containerssh/log"
	"github.com/containerssh/sshserver"
)

// New creates a new NetworkConnectionHandler for a specific client.
func New(client net.TCPAddr, connectionID string, config Config, logger log.Logger) (
	sshserver.NetworkConnectionHandler,
	error,
) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &networkHandler{
		mutex:               &sync.Mutex{},
		client:              client,
		connectionID:        connectionID,
		config:              config,
		logger:              logger,
		disconnected:        false,
		dockerClientFactory: &dockerV20ClientFactory{},
	}, nil
}
