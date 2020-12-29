package docker_test

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/containerssh/geoip"
	"github.com/containerssh/log"
	"github.com/containerssh/metrics"
	"github.com/containerssh/sshserver"
	"github.com/containerssh/structutils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/containerssh/docker"
)

func must(t *testing.T, arg bool) {
	if !arg {
		t.FailNow()
	}
}

func getDocker(t *testing.T, config docker.Config) (sshserver.NetworkConnectionHandler, string) {
	connectionID := sshserver.GenerateConnectionID()
	geoipProvider, err := geoip.New(geoip.Config{
		Provider: geoip.DummyProvider,
	})
	must(t, assert.NoError(t, err))
	collector := metrics.New(geoipProvider)
	dr, err := docker.New(
		net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 2222,
			Zone: "",
		},
		connectionID,
		config,
		createLogger(t),
		collector.MustCreateCounter("backend_requests", "", ""),
		collector.MustCreateCounter("backend_failures", "", ""),
	)
	must(t, assert.NoError(t, err))
	return dr, connectionID
}

//goland:noinspection GoDeprecation
func getDockerRun(t *testing.T, config docker.DockerRunConfig) sshserver.NetworkConnectionHandler {
	geoipProvider, err := geoip.New(geoip.Config{
		Provider: geoip.DummyProvider,
	})
	must(t, assert.NoError(t, err))
	collector := metrics.New(geoipProvider)
	dr, err := docker.NewDockerRun(
		net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 2222,
			Zone: "",
		},
		sshserver.GenerateConnectionID(),
		config,
		createLogger(t),
		collector.MustCreateCounter("backend_requests", "", ""),
		collector.MustCreateCounter("backend_failures", "", ""),
	)
	must(t, assert.NoError(t, err))
	return dr
}

func TestConnectAndDisconnectShouldCreateAndRemoveContainer(t *testing.T) {
	t.Parallel()

	config := docker.Config{}
	structutils.Defaults(&config)

	config.Execution.Launch.ContainerConfig.Image = "containerssh/containerssh-guest-image"

	dr, connectionID := getDocker(t, config)
	_, err := dr.OnHandshakeSuccess("test")
	defer dr.OnDisconnect()
	must(t, assert.Nil(t, err))

	dockerClient, err := client.NewClientWithOpts(
		client.WithHost(config.Connection.Host),
	)
	must(t, assert.Nil(t, err))
	dockerClient.NegotiateAPIVersion(context.Background())
	f := filters.NewArgs()
	f.Add("label", "containerssh_username=test")
	f.Add("label", "containerssh_ip=127.0.0.1")
	f.Add("label", "containerssh_connection_id="+connectionID)
	containers, err := dockerClient.ContainerList(
		context.Background(),
		types.ContainerListOptions{
			Filters: f,
		},
	)
	must(t, assert.Nil(t, err))
	must(t, assert.Equal(t, 1, len(containers)))
	must(t, assert.Equal(t, "running", containers[0].State))

	dr.OnDisconnect()
	_, err = dockerClient.ContainerInspect(context.Background(), containers[0].ID)
	must(t, assert.True(t, client.IsErrNotFound(err)))
}

func TestSingleSessionShouldRunProgram(t *testing.T) {
	t.Parallel()

	config := docker.Config{}
	structutils.Defaults(&config)

	dr, _ := getDocker(t, config)
	ssh, err := dr.OnHandshakeSuccess("test")
	must(t, assert.Nil(t, err))
	defer dr.OnDisconnect()

	session, err := ssh.OnSessionChannel(0, []byte{})
	must(t, assert.Nil(t, err))

	stdin := bytes.NewReader([]byte{})
	stdoutReader, stdout := io.Pipe()
	var stderrBytes bytes.Buffer
	stderr := bufio.NewWriter(&stderrBytes)
	done := make(chan struct{})
	status := 0
	go func() {
		assert.NoError(t, readUntil(stdoutReader, []byte("Hello world!\n")))
	}()

	err = session.OnExecRequest(
		0,
		"echo \"Hello world!\"",
		stdin,
		stdout,
		stderr,
		func(exitStatus sshserver.ExitStatus) {
			status = int(exitStatus)
			done <- struct{}{}
		},
	)
	must(t, assert.Nil(t, err))
	<-done
	assert.Equal(t, "", stderrBytes.String())
	assert.Equal(t, 0, status)
}

func TestSingleSessionShouldRunProgramDockerRunConfig(t *testing.T) {
	t.Parallel()

	//goland:noinspection GoDeprecation
	config := docker.DockerRunConfig{}
	structutils.Defaults(&config)
	testFile, err := os.Open("testdata/config-0.3.yaml")
	assert.NoError(t, err)
	unmarshaller := yaml.NewDecoder(testFile)
	unmarshaller.KnownFields(true)
	assert.NoError(t, unmarshaller.Decode(&config))

	dr := getDockerRun(t, config)
	ssh, err := dr.OnHandshakeSuccess("test")
	must(t, assert.Nil(t, err))
	defer dr.OnDisconnect()

	session, err := ssh.OnSessionChannel(0, []byte{})
	must(t, assert.Nil(t, err))

	stdin := bytes.NewReader([]byte{})
	stdoutReader, stdout := io.Pipe()
	var stderrBytes bytes.Buffer
	stderr := bufio.NewWriter(&stderrBytes)
	done := make(chan struct{})
	status := 0
	go func() {
		assert.NoError(t, readUntil(stdoutReader, []byte("1\n")))
	}()

	err = session.OnExecRequest(
		0,
		"echo $$",
		stdin,
		stdout,
		stderr,
		func(exitStatus sshserver.ExitStatus) {
			status = int(exitStatus)
			done <- struct{}{}
		},
	)
	must(t, assert.Nil(t, err))
	<-done
	assert.Equal(t, "", stderrBytes.String())
	assert.Equal(t, 0, status)
}

func TestSettingEnvShouldWork(t *testing.T) {
	t.Parallel()

	config := docker.Config{}
	structutils.Defaults(&config)

	dr, _ := getDocker(t, config)
	ssh, err := dr.OnHandshakeSuccess("test")
	must(t, assert.Nil(t, err))
	defer dr.OnDisconnect()

	session, err := ssh.OnSessionChannel(0, []byte{})
	must(t, assert.Nil(t, err))

	stdin := bytes.NewReader([]byte{})
	stdoutReader, stdout := io.Pipe()
	var stderrBytes bytes.Buffer
	stderr := bufio.NewWriter(&stderrBytes)
	done := make(chan struct{})
	status := 0

	assert.NoError(t, session.OnEnvRequest(0, "FOO", "bar"))

	go func() {
		assert.NoError(t, readUntil(stdoutReader, []byte("bar\n")))
	}()

	err = session.OnExecRequest(
		1,
		"echo \"$FOO\"",
		stdin,
		stdout,
		stderr,
		func(exitStatus sshserver.ExitStatus) {
			status = int(exitStatus)
			done <- struct{}{}
		},
	)
	must(t, assert.Nil(t, err))
	<-done
	assert.Equal(t, "", stderrBytes.String())
	assert.Equal(t, 0, status)
}

func TestSingleSessionShouldRunShell(t *testing.T) {
	t.Parallel()

	dr, ssh := initDockerRun(t)
	defer dr.OnDisconnect()

	var err error
	session, err := ssh.OnSessionChannel(0, []byte{})
	must(t, assert.Nil(t, err))

	stdin, stdinWriter := io.Pipe()
	stdoutReader, stdout := io.Pipe()
	_, stderr := io.Pipe()
	done := make(chan struct{})
	status := 0
	assert.NoError(t, session.OnEnvRequest(0, "foo", "bar"))
	assert.NoError(t, session.OnPtyRequest(1, "xterm", 80, 25, 800, 600, []byte{}))
	go func() {
		assert.NoError(t, readUntil(stdoutReader, []byte("# ")))

		assert.NoError(t, session.OnWindow(2, 120, 25, 800, 600))

		_, err = stdinWriter.Write([]byte("tput cols\n"))
		assert.NoError(t, readUntil(stdoutReader, []byte("tput cols\r\n120\r\n# ")))

		_, err = stdinWriter.Write([]byte("echo \"Hello world!\"\n"))
		assert.NoError(t, err)

		assert.NoError(t, readUntil(stdoutReader, []byte("echo \"Hello world!\"\r\nHello world!\r\n# ")))

		_, err = stdinWriter.Write([]byte("exit\n"))
		assert.NoError(t, err)

		assert.NoError(t, readUntil(stdoutReader, []byte("exit\r\n")))
	}()
	err = session.OnShell(
		3,
		stdin,
		stdout,
		stderr,
		func(exitStatus sshserver.ExitStatus) {
			status = int(exitStatus)
			done <- struct{}{}
		},
	)
	must(t, assert.Nil(t, err))
	<-done
	assert.Equal(t, 0, status)
}

func TestSendingSignalShouldWork(t *testing.T) {
	t.Parallel()

	config := docker.Config{}
	structutils.Defaults(&config)

	dr, _ := getDocker(t, config)
	ssh, err := dr.OnHandshakeSuccess("test")
	must(t, assert.Nil(t, err))
	defer dr.OnDisconnect()

	session, err := ssh.OnSessionChannel(0, []byte{})
	must(t, assert.Nil(t, err))

	stdin := bytes.NewReader([]byte{})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	done := make(chan struct{})
	status := 0

	assert.NoError(
		t,
		session.OnPtyRequest(0, "xterm", 80, 25, 800, 600, []byte("")),
	)

	go func() {
		time.Sleep(time.Second)
		assert.NoError(t, session.OnSignal(2, "USR1"))
	}()

	err = session.OnExecRequest(
		1,
		"sleep infinity & PID=$!; trap \"kill $PID\" USR1; wait; echo 'USR1 received'",
		stdin,
		&stdout,
		&stderr,
		func(exitStatus sshserver.ExitStatus) {
			status = int(exitStatus)
			done <- struct{}{}
		},
	)
	must(t, assert.Nil(t, err))
	<-done
	stderrBytes := stderr.Bytes()
	stdoutBytes := stdout.Bytes()
	assert.Equal(t, []byte(nil), stderrBytes)
	assert.Equal(t, []byte("USR1 received\r\n"), stdoutBytes)
	assert.Equal(t, 0, status)
}

func readUntil(reader io.Reader, buffer []byte) error {
	byteBuffer := bytes.NewBuffer([]byte{})
	for {
		buf := make([]byte, 1024)
		n, err := reader.Read(buf)
		if err != nil {
			return err
		}
		byteBuffer.Write(buf[:n])
		if bytes.Equal(byteBuffer.Bytes(), buffer) {
			return nil
		}
	}
}

func initDockerRun(t *testing.T) (sshserver.NetworkConnectionHandler, sshserver.SSHConnectionHandler) {
	config := docker.Config{}
	structutils.Defaults(&config)
	config.Execution.DisableAgent = true
	config.Execution.ShellCommand = []string{"/bin/sh"}

	dr, _ := getDocker(t, config)
	ssh, err := dr.OnHandshakeSuccess("test")
	must(t, assert.Nil(t, err))
	return dr, ssh
}

func createLogger(t *testing.T) log.Logger {
	logger, err := log.New(
		log.Config{
			Level:  log.LevelDebug,
			Format: "text",
		}, "docker", os.Stdout,
	)
	assert.Nil(t, err, "failed to create logger (%v)", err)
	return logger
}
