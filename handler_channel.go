package docker

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/containerssh/sshserver"
	"github.com/containerssh/unixutils"
)

type channelHandler struct {
	channelID      uint64
	networkHandler *networkHandler
	username       string
	env            map[string]string
	pty            bool
	columns        uint32
	rows           uint32
	exitSent       bool
	exec           dockerExecution
}

func (c *channelHandler) OnUnsupportedChannelRequest(_ uint64, _ string, _ []byte) {}

func (c *channelHandler) OnFailedDecodeChannelRequest(_ uint64, _ string, _ []byte, _ error) {}

func (c *channelHandler) OnEnvRequest(_ uint64, name string, value string) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec != nil {
		return fmt.Errorf("program already running")
	}
	c.env[name] = value
	return nil
}

func (c *channelHandler) OnPtyRequest(
	_ uint64,
	term string,
	columns uint32,
	rows uint32,
	_ uint32,
	_ uint32,
	_ []byte,
) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec != nil {
		return fmt.Errorf("program already running")
	}
	c.env["TERM"] = term
	c.rows = rows
	c.columns = columns
	c.pty = true
	return nil
}

func (c *channelHandler) parseProgram(program string) []string {
	programParts, err := unixutils.ParseCMD(program)
	if err != nil {
		return []string{"/bin/sh", "-c", program}
	} else {
		if strings.HasPrefix(programParts[0], "/") || strings.HasPrefix(
			programParts[0],
			"./",
		) || strings.HasPrefix(programParts[0], "../") {
			return programParts
		} else {
			return []string{"/bin/sh", "-c", program}
		}
	}
}

func (c *channelHandler) run(
	ctx context.Context,
	program []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	onExit func(exitStatus sshserver.ExitStatus),
) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec != nil {
		return fmt.Errorf("program already running")
	}

	var err error
	var realOnExit func(exitStatus sshserver.ExitStatus)
	switch c.networkHandler.config.Execution.Mode {
	case ExecutionModeConnection:
		realOnExit, err = c.handleExecModeConnection(ctx, program, onExit)
		if err != nil {
			return err
		}
	case ExecutionModeSession:
		realOnExit, err = c.handleExecModeSession(ctx, program, onExit)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid execution mode: %s", c.networkHandler.config.Execution.Mode)
	}

	go c.exec.run(
		stdout, stderr, stdin, func(exitStatus int) {
			c.networkHandler.mutex.Lock()
			defer c.networkHandler.mutex.Unlock()
			if c.exitSent {
				return
			}
			c.exitSent = true
			realOnExit(sshserver.ExitStatus(exitStatus))
		},
	)

	return nil
}

func (c *channelHandler) handleExecModeConnection(
	ctx context.Context,
	program []string,
	onExit func(exitStatus sshserver.ExitStatus),
) (func(exitStatus sshserver.ExitStatus), error) {
	exec, err := c.networkHandler.container.createExec(ctx, program, c.env, c.pty)
	if err != nil {
		return nil, err
	}
	c.exec = exec
	if c.pty {
		err := c.exec.resize(ctx, uint(c.rows), uint(c.columns))
		if err != nil {
			return nil, err
		}
	}
	return onExit, nil
}

func (c *channelHandler) handleExecModeSession(
	ctx context.Context,
	program []string,
	onExit func(exitStatus sshserver.ExitStatus),
) (func(exitStatus sshserver.ExitStatus), error) {
	cnt, err := c.networkHandler.dockerClient.createContainer(
		ctx,
		c.networkHandler.labels,
		c.env,
		&c.pty,
		program,
	)
	if err != nil {
		return nil, err
	}
	removeContainer := func() {
		ctx, cancelFunc := context.WithTimeout(
			context.Background(), c.networkHandler.config.Timeouts.ContainerStop,
		)
		defer cancelFunc()
		_ = cnt.remove(ctx)
	}
	c.exec, err = cnt.attach(ctx)
	if err != nil {
		removeContainer()
		return nil, err
	}
	if c.pty {
		err := c.exec.resize(ctx, uint(c.rows), uint(c.columns))
		if err != nil {
			removeContainer()
			return nil, err
		}
	}
	if err := cnt.start(ctx); err != nil {
		removeContainer()
		return nil, err
	}
	onExitWrapper := func(exitStatus sshserver.ExitStatus) {
		onExit(exitStatus)
		removeContainer()
	}
	return onExitWrapper, nil
}

func (c *channelHandler) OnExecRequest(
	_ uint64,
	program string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	onExit func(exitStatus sshserver.ExitStatus),
) error {
	if c.networkHandler.config.Execution.disableCommand {
		return fmt.Errorf("command execution is disabled")
	}
	startContext, cancelFunc := context.WithTimeout(context.Background(), c.networkHandler.config.Timeouts.CommandStart)
	defer cancelFunc()
	return c.run(startContext, c.parseProgram(program), stdin, stdout, stderr, onExit)
}

func (c *channelHandler) OnShell(
	_ uint64,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	onExit func(exitStatus sshserver.ExitStatus),
) error {
	startContext, cancelFunc := context.WithTimeout(context.Background(), c.networkHandler.config.Timeouts.CommandStart)
	defer cancelFunc()

	return c.run(startContext, c.getDefaultShell(), stdin, stdout, stderr, onExit)
}

func (c *channelHandler) getDefaultShell() []string {
	return c.networkHandler.config.Execution.ShellCommand
}

func (c *channelHandler) OnSubsystem(
	_ uint64,
	subsystem string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	onExit func(exitStatus sshserver.ExitStatus),
) error {
	startContext, cancelFunc := context.WithTimeout(context.Background(), c.networkHandler.config.Timeouts.CommandStart)
	defer cancelFunc()

	if binary, ok := c.networkHandler.config.Execution.Subsystems[subsystem]; ok {
		return c.run(startContext, []string{binary}, stdin, stdout, stderr, onExit)
	}
	return fmt.Errorf("subsystem not supported")
}

func (c *channelHandler) OnSignal(_ uint64, _ string) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec == nil {
		return fmt.Errorf("program not running")
	}

	return nil
}

func (c *channelHandler) OnWindow(_ uint64, columns uint32, rows uint32, _ uint32, _ uint32) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec == nil {
		return fmt.Errorf("program not running")
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelFunc()

	return c.exec.resize(ctx, uint(rows), uint(columns))
}
