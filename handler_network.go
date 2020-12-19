package docker

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/containerssh/log"
	"github.com/containerssh/sshserver"
)

type networkHandler struct {
	mutex               *sync.Mutex
	client              net.TCPAddr
	username            string
	connectionID        string
	config              Config
	container           dockerContainer
	dockerClient        dockerClient
	dockerClientFactory dockerClientFactory
	logger              log.Logger
	disconnected        bool
	labels              map[string]string
}

func (n *networkHandler) OnAuthPassword(_ string, _ []byte) (response sshserver.AuthResponse, reason error) {
	return sshserver.AuthResponseUnavailable, fmt.Errorf("docker does not support authentication")
}

func (n *networkHandler) OnAuthPubKey(_ string, _ string) (response sshserver.AuthResponse, reason error) {
	return sshserver.AuthResponseUnavailable, fmt.Errorf("docker does not support authentication")
}

func (n *networkHandler) OnHandshakeFailed(_ error) {}

func (n *networkHandler) OnHandshakeSuccess(username string) (
	connection sshserver.SSHConnectionHandler,
	failureReason error,
) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	ctx, cancelFunc := context.WithTimeout(context.Background(), n.config.Timeouts.ContainerStart)
	defer cancelFunc()
	n.username = username

	if err := n.setupDockerClient(ctx, n.config); err != nil {
		return nil, err
	}
	if err := n.pullImage(ctx); err != nil {
		return nil, err
	}
	labels := map[string]string{}
	labels["containerssh_connection_id"] = n.connectionID
	labels["containerssh_ip"] = n.client.IP.String()
	labels["containerssh_username"] = n.username
	n.labels = labels
	var cnt dockerContainer
	var err error
	if n.config.Execution.Mode == ExecutionModeConnection {
		if cnt, err = n.dockerClient.createContainer(ctx, labels, nil, nil, nil); err != nil {
			return nil, err
		}
		n.container = cnt
		if err := n.container.start(ctx); err != nil {
			return nil, err
		}
	}

	return &sshConnectionHandler{
		networkHandler: n,
		username:       username,
	}, nil
}

func (n *networkHandler) pullNeeded(ctx context.Context) (bool, error) {
	switch n.config.Execution.ImagePullPolicy {
	case ImagePullPolicyNever:
		return false, nil
	case ImagePullPolicyAlways:
		return true, nil
	}

	image := n.dockerClient.getImageName()
	if !strings.Contains(image, ":") || strings.HasSuffix(image, ":latest") {
		return true, nil
	}

	hasImage, err := n.dockerClient.hasImage(ctx)
	if err != nil {
		return true, err
	}
	return !hasImage, nil
}

func (n *networkHandler) pullImage(ctx context.Context) (err error) {
	pullNeeded, err := n.pullNeeded(ctx)
	if err != nil || !pullNeeded {
		return err
	}

	return n.dockerClient.pullImage(ctx)
}

func (n *networkHandler) setupDockerClient(ctx context.Context, config Config) error {
	if n.dockerClient == nil {
		dockerClient, err := n.dockerClientFactory.get(ctx, config, n.logger)
		if err != nil {
			return fmt.Errorf("failed to create Docker client (%w)", err)
		}
		n.dockerClient = dockerClient
	}
	return nil
}

func (n *networkHandler) OnDisconnect() {
	n.disconnected = true
	ctx, cancelFunc := context.WithTimeout(context.Background(), n.config.Timeouts.ContainerStop)
	defer cancelFunc()
	n.mutex.Lock()
	if n.container != nil {
		_ = n.container.remove(ctx)
		n.mutex.Unlock()
	}
}
