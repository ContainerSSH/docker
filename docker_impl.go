package docker

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/containerssh/log"
	"github.com/containerssh/structutils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type dockerV20ClientFactory struct {
}

func (f *dockerV20ClientFactory) getDockerClient(ctx context.Context, config Config) (*client.Client, error) {
	httpClient, err := getHTTPClient(config)
	if err != nil {
		return nil, err
	}
	cli, err := client.NewClientWithOpts(
		client.WithHost(config.Connection.Host),
		client.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, err
	}
	cli.NegotiateAPIVersion(ctx)
	return cli, nil
}

func (f *dockerV20ClientFactory) get(ctx context.Context, config Config, logger log.Logger) (dockerClient, error) {
	if config.Execution.Launch.ContainerConfig == nil || config.Execution.Launch.ContainerConfig.Image == "" {
		return nil, fmt.Errorf("no image name specified")
	}

	dockerClient, err := f.getDockerClient(ctx, config)
	if err != nil {
		return nil, err
	}

	return &dockerV20Client{
		config:       config,
		dockerClient: dockerClient,
		logger:       logger,
	}, nil
}

type dockerV20Client struct {
	config       Config
	dockerClient *client.Client
	logger       log.Logger
}

func (d *dockerV20Client) getImageName() string {
	return d.config.Execution.Launch.ContainerConfig.Image
}

func (d *dockerV20Client) hasImage(ctx context.Context) (bool, error) {
	image := d.config.Execution.Launch.ContainerConfig.Image
	d.logger.Debugf(
		"Checking if image %s exists locally...", image,
	)
	var lastError error
loop:
	for {
		_, _, lastError := d.dockerClient.ImageInspectWithRaw(ctx, image)
		if lastError != nil {
			if client.IsErrNotFound(lastError) {
				return false, nil
			}
			d.logger.Warninge(
				fmt.Errorf("failed to list images, retrying in 10 seconds (%w)", lastError),
			)
		} else {
			return true, nil
		}
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	d.logger.Errore(
		fmt.Errorf("failed to list images, giving up (%w)", lastError),
	)

	return false, lastError
}

func (d *dockerV20Client) pullImage(ctx context.Context) error {
	image, err := getCanonicalImageName(d.config.Execution.Launch.ContainerConfig.Image)
	if err != nil {
		return err
	}

	d.logger.Debugf(
		"Pulling image %s...", image,
	)
	var lastError error
loop:
	for {
		var pullReader io.ReadCloser
		pullReader, lastError = d.dockerClient.ImagePull(ctx, image, types.ImagePullOptions{})
		if lastError == nil {
			_, lastError = ioutil.ReadAll(pullReader)
			if lastError == nil {
				lastError = pullReader.Close()
				if lastError == nil {
					return nil
				}
			}
		}
		if pullReader != nil {
			_ = pullReader.Close()
		}
		d.logger.Warninge(
			fmt.Errorf("failed to pull image %s, retrying in 10 seconds (%w)", image, lastError),
		)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err = fmt.Errorf("failed to pull image %s, giving up (%w)", image, lastError)
	d.logger.Errore(
		err,
	)
	return err
}

func (d *dockerV20Client) createContainer(
	ctx context.Context,
	labels map[string]string,
	env map[string]string,
	tty *bool,
	cmd []string,
) (dockerContainer, error) {
	d.logger.Debugf(
		"Creating container...",
	)
	containerConfig := d.config.Execution.Launch.ContainerConfig
	newConfig, err := d.createConfig(containerConfig, labels, env, tty, cmd)
	if err != nil {
		return nil, err
	}

	var lastError error
loop:
	for {
		var body container.ContainerCreateCreatedBody
		body, lastError = d.dockerClient.ContainerCreate(
			ctx,
			newConfig,
			d.config.Execution.Launch.HostConfig,
			d.config.Execution.Launch.NetworkConfig,
			d.config.Execution.Launch.Platform,
			d.config.Execution.Launch.ContainerName,
		)
		if lastError == nil {
			return &dockerV20Container{
				config:       d.config,
				containerID:  body.ID,
				dockerClient: d.dockerClient,
				logger:       d.logger,
				tty:          d.config.Execution.Launch.ContainerConfig.Tty,
			}, nil
		}
		d.logger.Warninge(fmt.Errorf("failed to create container, retrying in 10 seconds (%w)", lastError))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err = fmt.Errorf("failed to create container, giving up (%w)", lastError)
	d.logger.Errore(err)
	return nil, err
}

func (d *dockerV20Client) createConfig(
	containerConfig *container.Config,
	labels map[string]string,
	env map[string]string,
	tty *bool,
	cmd []string,
) (*container.Config, error) {
	newConfig := &container.Config{}
	if containerConfig != nil {
		if err := structutils.Copy(newConfig, containerConfig); err != nil {
			return nil, err
		}
	}
	if newConfig.Labels == nil {
		newConfig.Labels = map[string]string{}
	}
	newConfig.Cmd = d.config.Execution.IdleCommand
	for k, v := range labels {
		newConfig.Labels[k] = v
	}

	newConfig.Env = append(newConfig.Env, createEnv(env)...)
	if tty != nil {
		newConfig.Tty = *tty
		newConfig.AttachStdin = true
		newConfig.AttachStdout = true
		newConfig.AttachStderr = true
		newConfig.OpenStdin = true
		newConfig.StdinOnce = true
		newConfig.Cmd = cmd
	}
	return newConfig, nil
}

type dockerV20Container struct {
	config       Config
	containerID  string
	logger       log.Logger
	dockerClient *client.Client
	tty          bool
}

func (d *dockerV20Container) attach(ctx context.Context) (dockerExecution, error) {
	d.logger.Debugf(
		"Attaching to container...",
	)
	var attachResult types.HijackedResponse
	var lastError error
loop:
	for {
		attachResult, lastError = d.dockerClient.ContainerAttach(
			ctx,
			d.containerID,
			types.ContainerAttachOptions{
				Stream: true,
				Stdin:  true,
				Stdout: true,
				Stderr: true,
				Logs:   true,
			},
		)
		if lastError == nil {
			return &dockerV20Exec{
				container:    d,
				execID:       "",
				dockerClient: d.dockerClient,
				logger:       d.logger,
				attachResult: attachResult,
				tty:          d.tty,
				pid:          1,
			}, nil
		}
		d.logger.Warninge(fmt.Errorf("failed to attach to exec, retrying in 10 seconds (%w)", lastError))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err := fmt.Errorf("failed to attach to exec, giving up (%w)", lastError)
	d.logger.Errore(err)
	return nil, err
}

func (d *dockerV20Container) start(ctx context.Context) error {
	d.logger.Debugf(
		"Starting to container...",
	)
	var lastError error
loop:
	for {
		lastError = d.dockerClient.ContainerStart(
			ctx,
			d.containerID,
			types.ContainerStartOptions{},
		)
		if lastError == nil {
			return nil
		}
		d.logger.Warninge(fmt.Errorf("failed to start container, retrying in 10 seconds (%w)", lastError))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err := fmt.Errorf("failed to start container, giving up (%w)", lastError)
	d.logger.Errore(err)
	return err
}

func (d *dockerV20Container) remove(ctx context.Context) error {
	d.logger.Debugf(
		"Removing container...",
	)
	var lastError error
loop:
	for {
		_, lastError = d.dockerClient.ContainerInspect(ctx, d.containerID)
		if lastError != nil && client.IsErrNotFound(lastError) {
			return nil
		}

		if lastError == nil {
			lastError = d.dockerClient.ContainerRemove(
				ctx, d.containerID, types.ContainerRemoveOptions{
					Force: true,
				},
			)
			if lastError == nil {
				return nil
			}
		}
		d.logger.Warninge(
			fmt.Errorf("failed to remove container on disconnect, retrying in 10 seconds (%w)", lastError),
		)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err := fmt.Errorf("failed to remove container on disconnect, giving up (%w)", lastError)
	d.logger.Errore(
		err,
	)
	return err
}

func (d *dockerV20Container) createExec(
	ctx context.Context,
	program []string,
	env map[string]string,
	tty bool,
) (dockerExecution, error) {
	d.logger.Debugf(
		"Creating exec...",
	)
	execConfig := d.createExecConfig(env, tty, program)
	execID, err := d.realCreateExec(ctx, execConfig)
	if err != nil {
		return nil, err
	}

	attachResult, err := d.attachExec(ctx, execID, execConfig)
	if err != nil {
		return nil, err
	}

	pid := -1
	if !d.config.Execution.DisableAgent {
		// Read PID from execution
		pidBytes := make([]byte, 4)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			_, err := attachResult.Reader.Read(pidBytes)
			if err != nil {
				return nil, err
			}
			pid = int(binary.LittleEndian.Uint32(pidBytes))
		}
	}

	return &dockerV20Exec{
		container:    d,
		execID:       execID,
		dockerClient: d.dockerClient,
		logger:       d.logger,
		attachResult: attachResult,
		tty:          tty,
		pid:          pid,
	}, nil
}

func (d *dockerV20Container) realCreateExec(ctx context.Context, execConfig types.ExecConfig) (string, error) {
	var lastError error
loop:
	for {
		var response types.IDResponse
		response, lastError = d.dockerClient.ContainerExecCreate(
			ctx,
			d.containerID,
			execConfig,
		)
		if lastError == nil {
			return response.ID, nil
		}
		d.logger.Warninge(fmt.Errorf("failed to create exec, retrying in 10 seconds (%w)", lastError))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err := fmt.Errorf("failed to create exec, giving up (%w)", lastError)
	d.logger.Errore(err)
	return "", err
}

func (d *dockerV20Container) createExecConfig(env map[string]string, tty bool, program []string) types.ExecConfig {
	dockerEnv := createEnv(env)
	if !d.config.Execution.DisableAgent {
		agentPrefix := []string{
			d.config.Execution.AgentPath,
			"console",
			"--pid",
			"--",
		}
		program = append(agentPrefix, program...)
	}
	execConfig := types.ExecConfig{
		Tty:          tty,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Env:          dockerEnv,
		Cmd:          program,
	}
	return execConfig
}

func createEnv(env map[string]string) []string {
	var dockerEnv []string
	for k, v := range env {
		dockerEnv = append(dockerEnv, fmt.Sprintf("%s=%s", k, v))
	}
	return dockerEnv
}

func (d *dockerV20Container) attachExec(ctx context.Context, execID string, config types.ExecConfig) (types.HijackedResponse, error) {
	d.logger.Debugf(
		"Attaching exec...",
	)
	var attachResult types.HijackedResponse
	var lastError error
loop:
	for {
		attachResult, lastError = d.dockerClient.ContainerExecAttach(
			ctx,
			execID,
			types.ExecStartCheck{
				Detach: false,
				Tty:    config.Tty,
			},
		)
		if lastError == nil {
			return attachResult, nil
		}
		if isPermanentError(lastError) {
			err := fmt.Errorf("failed to attach to exec, permanent error (%w)", lastError)
			d.logger.Errore(err)
			return types.HijackedResponse{}, err
		}
		d.logger.Warninge(fmt.Errorf("failed to attach to exec, retrying in 10 seconds (%w)", lastError))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err := fmt.Errorf("failed to attach to exec, giving up (%w)", lastError)
	d.logger.Errore(err)
	return types.HijackedResponse{}, err
}

type dockerV20Exec struct {
	container    *dockerV20Container
	execID       string
	dockerClient *client.Client
	logger       log.Logger
	attachResult types.HijackedResponse
	tty          bool
	pid          int
}

var cannotSendSignalError = errors.New("cannot send signal")

func (d *dockerV20Exec) signal(ctx context.Context, sig string) error {
	if d.pid <= 0 {
		return cannotSendSignalError
	}
	if d.pid == 1 {
		return d.sendSignalToContainer(ctx, sig)
	}
	return d.sendSignalToProcess(ctx, sig)
}

func (d *dockerV20Exec) sendSignalToProcess(ctx context.Context, sig string) error {
	if d.container.config.Execution.DisableAgent {
		return fmt.Errorf("cannot send signal")
	}
	d.logger.Debugf("Using the exec facility to send signal %s to pid %d...", sig, d.pid)
	exec, err := d.container.createExec(
		ctx, []string{
			d.container.config.Execution.AgentPath,
			"signal",
			"--pid",
			strconv.Itoa(d.pid),
			"--signal",
			sig,
		}, map[string]string{}, false,
	)
	if err != nil {
		d.logger.Errorf(
			"cannot send %s signal to container %s pid %d (%v)",
			sig, d.container.containerID, d.pid, err,
		)
		return cannotSendSignalError
	}
	var stdoutBytes bytes.Buffer
	var stderrBytes bytes.Buffer
	stdin, stdinWriter := io.Pipe()
	done := make(chan struct{})
	exec.run(
		&stdoutBytes, &stderrBytes, stdin, func(exitStatus int) {
			if exitStatus != 0 {
				err = cannotSendSignalError
				d.logger.Errorf(
					"cannot send %s signal to container %s pid %d (%s)",
					sig, d.container.containerID, d.pid, stderrBytes,
				)
			}
			done <- struct{}{}
		},
	)
	<-done
	_ = stdinWriter.Close()
	return err
}

func (d *dockerV20Exec) sendSignalToContainer(ctx context.Context, sig string) error {
	var lastError error
loop:
	for {
		lastError = d.dockerClient.ContainerKill(ctx, d.container.containerID, sig)
		if lastError == nil {
			return nil
		}
		if isPermanentError(lastError) {
			err := fmt.Errorf(
				"cannot send %s signal to container %s, permanent error (%w)",
				sig, d.container.containerID, lastError,
			)
			d.logger.Errore(err)
			return err
		}
		d.logger.Warninge(
			fmt.Errorf(
				"cannot send %s signal to container %s, retrying in 10 seconds (%w)",
				sig, d.container.containerID, lastError,
			),
		)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	d.logger.Errorf("cannot send signal %s to container %s, giving up (%v)", sig, d.container.containerID, lastError)
	return cannotSendSignalError
}

func (d *dockerV20Exec) resize(ctx context.Context, height uint, width uint) error {
	d.logger.Debugf(
		"Resizing...",
	)
	var lastError error
loop:
	for {
		lastError = d.dockerClient.ContainerExecResize(
			ctx, d.execID, types.ResizeOptions{
				Height: height,
				Width:  width,
			},
		)
		if lastError == nil {
			return nil
		}
		if isPermanentError(lastError) {
			err := fmt.Errorf("failed to resize window, permanent error (%w)", lastError)
			// Debug level because resizes can fail for legitimate reasons, such as invalid program paths.
			d.logger.Debuge(err)
			return err
		}
		d.logger.Warninge(fmt.Errorf("failed to resize window, retrying in 10 seconds (%w)", lastError))
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err := fmt.Errorf("failed to resize exec, giving up (%w)", lastError)
	d.logger.Errore(err)
	return err
}

func (d *dockerV20Exec) run(stdout io.Writer, stderr io.Writer, stdin io.Reader, onExit func(exitStatus int)) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	if d.tty {
		go func() {
			defer d.done(onExit)
			_, err := io.Copy(stdout, d.attachResult.Reader)
			if err != nil && !errors.Is(err, io.EOF) {
				d.logger.Warninge(
					fmt.Errorf("failed to stream TTY output (%w)", err),
				)
			}
		}()
	} else {
		go func() {
			defer d.done(onExit)
			_, err := stdcopy.StdCopy(stdout, stderr, d.attachResult.Reader)
			if err != nil && !errors.Is(err, io.EOF) {
				d.logger.Warninge(
					fmt.Errorf("failed to stream raw output (%w)", err),
				)
			}
		}()
	}
	go func() {
		_, err := io.Copy(d.attachResult.Conn, stdin)
		if err != nil && !errors.Is(err, io.EOF) {
			d.logger.Warninge(
				fmt.Errorf("failed to stream input (%w)", err),
			)
		}
	}()
}

func (d *dockerV20Exec) done(onExit func(exitStatus int)) {
	d.pid = -1
	ctx, cancelFunc := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelFunc()
	var lastError error
loop:
	for {
		if d.execID != "" {
			var inspectResult types.ContainerExecInspect
			inspectResult, lastError = d.dockerClient.ContainerExecInspect(ctx, d.execID)
			if lastError == nil {
				if inspectResult.ExitCode < 0 {
					lastError = fmt.Errorf("negative exit code: %d", inspectResult.ExitCode)
				} else {
					onExit(inspectResult.ExitCode)
					return
				}
			}
		} else {
			var inspectResult types.ContainerJSON
			if err := d.stopContainer(ctx); err != nil {
				onExit(137)
				return
			}

			inspectResult, lastError = d.dockerClient.ContainerInspect(ctx, d.container.containerID)
			if lastError == nil {
				if inspectResult.State.Running {
					lastError = fmt.Errorf("container still running")
				} else if inspectResult.State.Restarting {
					lastError = fmt.Errorf("container restarting")
				} else if inspectResult.State.ExitCode < 0 {
					lastError = fmt.Errorf("negative exit code: %d", inspectResult.State.ExitCode)
				} else {
					onExit(inspectResult.State.ExitCode)
					return
				}
			}
		}
		if isPermanentError(lastError) {
			err := fmt.Errorf("failed to fetch exit code, permanent error (%w)", lastError)
			d.logger.Errore(err)
			return
		}
		d.logger.Warninge(
			fmt.Errorf("failed to fetch container exit code, retrying in 10 seconds (%w)", lastError),
		)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err := fmt.Errorf("failed to fetch container exit code, giving up (%w)", lastError)
	d.logger.Errore(err)
}

func (d *dockerV20Exec) stopContainer(ctx context.Context) error {
	var lastError error
loop:
	for {
		var inspectResult types.ContainerJSON
		inspectResult, lastError = d.dockerClient.ContainerInspect(ctx, d.container.containerID)
		if lastError == nil {
			if inspectResult.State.Status == "stopped" {
				return nil
			}
			d.logger.Debugf("Stopping container...")
			lastError = d.dockerClient.ContainerStop(
				ctx,
				d.container.containerID,
				&d.container.config.Timeouts.ContainerStop)
			if lastError == nil {
				return nil
			}
		}
		d.logger.Warninge(
			fmt.Errorf("failed to stop container, retrying in 10 seconds (%w)", lastError),
		)
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(10 * time.Second):
		}
	}
	if lastError == nil {
		lastError = fmt.Errorf("timeout")
	}
	err := fmt.Errorf("failed to stop container, giving up (%w)", lastError)
	d.logger.Errore(err)
	return err
}

func isPermanentError(err error) bool {
	return client.IsErrNotFound(err) ||
		client.IsErrNotImplemented(err) ||
		client.IsErrPluginPermissionDenied(err) ||
		client.IsErrUnauthorized(err) ||
		strings.Contains(err.Error(), "")
}
