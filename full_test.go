package docker_test

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/containerssh/geoip"
	"github.com/containerssh/log"
	"github.com/containerssh/metrics"
	"github.com/containerssh/service"
	"github.com/containerssh/sshserver"
	"github.com/containerssh/structutils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"

	"github.com/containerssh/docker"
)

func TestFullSSHServer(t *testing.T) {
	lifecycle, listen, err := createSSHServer()
	if !assert.NoError(t, err) {
		return
	}

	running := make(chan struct{})
	lifecycle.OnRunning(
		func(s service.Service, l service.Lifecycle) {
			running <- struct{}{}
		},
	)
	go func() {
		_ = lifecycle.Run()
	}()
	<-running
	defer lifecycle.Stop(context.Background())

	clientConfig := ssh.ClientConfig{
		User:            "test",
		Auth:            []ssh.AuthMethod{ssh.Password("")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshConnection, err := ssh.Dial("tcp", listen, &clientConfig)
	if !assert.NoError(t, err) {
		return
	}
	session, err := sshConnection.NewSession()
	if !assert.NoError(t, err) {
		return
	}
	output, err := session.CombinedOutput("echo \"Hello world!\"")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, []byte("Hello world!\n"), output)
}

func createSSHServer() (service.Lifecycle, string, error) {
	logger, err := log.New(
		log.Config{
			Level:  log.LevelDebug,
			Format: log.FormatText,
		},
		"ssh",
		os.Stdout,
	)
	if err != nil {
		return nil, "", err
	}
	config := sshserver.Config{}
	structutils.Defaults(&config)
	if err := config.GenerateHostKey(); err != nil {
		return nil, "", err
	}
	geo, err := geoip.New(
		geoip.Config{
			Provider: geoip.DummyProvider,
		},
	)
	if err != nil {
		return nil, "", err
	}
	metricsCollector := metrics.New(geo)
	srv, err := sshserver.New(
		config,
		&fullHandler{
			logger,
			metricsCollector.MustCreateCounter("backend_requests", "", ""),
			metricsCollector.MustCreateCounter("backend_errors", "", ""),
		},
		logger,
	)
	if err != nil {
		return nil, "", err
	}
	lifecycle := service.NewLifecycle(srv)
	listen := config.Listen
	return lifecycle, listen, err
}

type fullHandler struct {
	logger         log.Logger
	requestsMetric metrics.SimpleCounter
	errorsMetric   metrics.SimpleCounter
}

func (f *fullHandler) OnReady() error {
	return nil
}

func (f *fullHandler) OnShutdown(_ context.Context) {}

func (f *fullHandler) OnNetworkConnection(client net.TCPAddr, connectionID string) (
	sshserver.NetworkConnectionHandler,
	error,
) {
	config := docker.Config{}
	structutils.Defaults(&config)

	backend, err := docker.New(
		client,
		connectionID,
		config,
		f.logger,
		f.requestsMetric,
		f.errorsMetric,
	)
	if err != nil {
		return nil, err
	}
	return &nullAuthenticator{
		backend: backend,
	}, nil
}

type nullAuthenticator struct {
	backend sshserver.NetworkConnectionHandler
}

func (n *nullAuthenticator) OnAuthPassword(_ string, _ []byte) (
	response sshserver.AuthResponse,
	reason error,
) {
	return sshserver.AuthResponseSuccess, nil
}

func (n *nullAuthenticator) OnAuthPubKey(_ string, _ string) (
	response sshserver.AuthResponse,
	reason error,
) {
	return sshserver.AuthResponseSuccess, nil
}

func (n *nullAuthenticator) OnHandshakeFailed(_ error) {

}

func (n *nullAuthenticator) OnHandshakeSuccess(username string) (
	connection sshserver.SSHConnectionHandler,
	failureReason error,
) {
	return n.backend.OnHandshakeSuccess(username)
}

func (n *nullAuthenticator) OnDisconnect() {
	n.backend.OnDisconnect()
}
