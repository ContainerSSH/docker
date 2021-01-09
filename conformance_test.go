package docker_test

import (
	"net"
	"os"
	"testing"

	"github.com/containerssh/geoip"
	"github.com/containerssh/log"
	"github.com/containerssh/metrics"
	"github.com/containerssh/sshserver"
	"github.com/containerssh/structutils"
	"gopkg.in/yaml.v3"

	"github.com/containerssh/docker"
)

func TestConformance(t *testing.T) {
	var factories = map[string]func() (sshserver.NetworkConnectionHandler, error) {
		"dockerrun": func() (sshserver.NetworkConnectionHandler, error) {
			config := docker.DockerRunConfig{}
			structutils.Defaults(&config)
			testFile, err := os.Open("testdata/config-0.3.yaml")
			if err != nil {
				return nil, err
			}
			unmarshaller := yaml.NewDecoder(testFile)
			unmarshaller.KnownFields(true)
			if err := unmarshaller.Decode(&config); err != nil {
				return nil, err
			}

			return getDockerRun(config)
		},
		"session": func() (sshserver.NetworkConnectionHandler, error) {
			config := docker.Config{}
			structutils.Defaults(&config)
			config.Execution.ShellCommand = []string{"/bin/sh"}

			config.Execution.Mode = docker.ExecutionModeSession
			return getDocker(config)
		},
		"connection": func() (sshserver.NetworkConnectionHandler, error) {
			config := docker.Config{}
			structutils.Defaults(&config)
			config.Execution.ShellCommand = []string{"/bin/sh"}

			config.Execution.Mode = docker.ExecutionModeConnection
			return getDocker(config)
		},
	}

	sshserver.RunConformanceTests(t, factories)
}


func getDocker(config docker.Config) (sshserver.NetworkConnectionHandler, error) {
	connectionID := sshserver.GenerateConnectionID()
	geoipProvider, err := geoip.New(geoip.Config{
		Provider: geoip.DummyProvider,
	})
	if err != nil {
		return nil, err
	}
	collector := metrics.New(geoipProvider)
	logger, err := log.New(
		log.Config{
			Level:  log.LevelDebug,
			Format: log.FormatText,
		},
		"docker",
		os.Stdout,
	)
	if err != nil {
		return nil, err
	}
	return docker.New(
		net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 2222,
			Zone: "",
		},
		connectionID,
		config,
		logger,
		collector.MustCreateCounter("backend_requests", "", ""),
		collector.MustCreateCounter("backend_failures", "", ""),
	)
}

//goland:noinspection GoDeprecation
func getDockerRun(config docker.DockerRunConfig) (sshserver.NetworkConnectionHandler, error) {
	geoipProvider, err := geoip.New(geoip.Config{
		Provider: geoip.DummyProvider,
	})
	if err != nil {
		return nil, err
	}
	collector := metrics.New(geoipProvider)
	logger, err := log.New(
		log.Config{
			Level:  log.LevelDebug,
			Format: log.FormatText,
		},
		"dockerrun",
		os.Stdout,
	)
	if err != nil {
		return nil, err
	}
	return docker.NewDockerRun(
		net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 2222,
			Zone: "",
		},
		sshserver.GenerateConnectionID(),
		config,
		logger,
		collector.MustCreateCounter("backend_requests", "", ""),
		collector.MustCreateCounter("backend_failures", "", ""),
	)
}
