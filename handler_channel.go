package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/containerssh/sshserver"
	"github.com/containerssh/unixutils"
)

type channelHandler struct {
	sshserver.AbstractSessionChannelHandler

	channelID      uint64
	networkHandler *networkHandler
	username       string
	env            map[string]string
	pty            bool
	columns        uint32
	rows           uint32
	exitSent       bool
	exec           dockerExecution
	session        sshserver.SessionChannel
}

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
) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec != nil {
		return fmt.Errorf("program already running")
	}

	var err error
	switch c.networkHandler.config.Execution.Mode {
	case ExecutionModeConnection:
		err = c.handleExecModeConnection(ctx, program)
	case ExecutionModeSession:
		err = c.handleExecModeSession(ctx, program)
	default:
		return fmt.Errorf("invalid execution mode: %s", c.networkHandler.config.Execution.Mode)
	}
	if err != nil {
		return err
	}

	go c.exec.run(
		c.session.Stdin(),
		c.session.Stdout(),
		c.session.Stderr(),
		c.session.CloseWrite,
		func(exitStatus int) {
			c.session.ExitStatus(uint32(exitStatus))
			_ = c.session.Close()
		},
	)

	return nil
}

func (c *channelHandler) handleExecModeConnection(
	ctx context.Context,
	program []string,
) error {
	exec, err := c.networkHandler.container.createExec(ctx, program, c.env, c.pty)
	if err != nil {
		return err
	}
	c.exec = exec
	if c.pty {
		_ = c.exec.resize(ctx, uint(c.rows), uint(c.columns))
	}
	return nil
}

func (c *channelHandler) handleExecModeSession(
	ctx context.Context,
	program []string,
) error {
	cnt, err := c.networkHandler.dockerClient.createContainer(
		ctx,
		c.networkHandler.labels,
		c.env,
		&c.pty,
		program,
	)
	if err != nil {
		return  err
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
		return err
	}
	if err := cnt.start(ctx); err != nil {
		removeContainer()
		return err
	}
	if c.pty {
		err := c.exec.resize(ctx, uint(c.rows), uint(c.columns))
		if err != nil {
			removeContainer()
			return err
		}
	}
	return nil
}

func (c *channelHandler) OnExecRequest(
	_ uint64,
	program string,
) error {
	if c.networkHandler.config.Execution.disableCommand {
		return fmt.Errorf("command execution is disabled")
	}
	startContext, cancelFunc := context.WithTimeout(context.Background(), c.networkHandler.config.Timeouts.CommandStart)
	defer cancelFunc()
	return c.run(
		startContext,
		c.parseProgram(program),
	)
}

func (c *channelHandler) OnShell(
	_ uint64,
) error {
	startContext, cancelFunc := context.WithTimeout(context.Background(), c.networkHandler.config.Timeouts.CommandStart)
	defer cancelFunc()

	return c.run(startContext, c.getDefaultShell())
}

func (c *channelHandler) getDefaultShell() []string {
	return c.networkHandler.config.Execution.ShellCommand
}

func (c *channelHandler) OnSubsystem(
	_ uint64,
	subsystem string,
) error {
	startContext, cancelFunc := context.WithTimeout(context.Background(), c.networkHandler.config.Timeouts.CommandStart)
	defer cancelFunc()

	if binary, ok := c.networkHandler.config.Execution.Subsystems[subsystem]; ok {
		return c.run(startContext, []string{binary})
	}
	return fmt.Errorf("subsystem not supported")
}

func (c *channelHandler) OnSignal(_ uint64, signal string) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec == nil {
		return fmt.Errorf("program not running")
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), c.networkHandler.config.Timeouts.Signal)
	defer cancelFunc()

	return c.exec.signal(ctx, signal)
}

func (c *channelHandler) OnWindow(_ uint64, columns uint32, rows uint32, _ uint32, _ uint32) error {
	c.networkHandler.mutex.Lock()
	defer c.networkHandler.mutex.Unlock()
	if c.exec == nil {
		return fmt.Errorf("program not running")
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), c.networkHandler.config.Timeouts.Window)
	defer cancelFunc()

	return c.exec.resize(ctx, uint(rows), uint(columns))
}

func (c *channelHandler) OnClose() {
	if c.exec != nil {
		c.exec.kill()
	}
}

func (c *channelHandler) OnShutdown(shutdownContext context.Context) {
	if c.exec != nil {
		c.exec.term(shutdownContext)
		// We wait for the program to exit. This is not needed in session or connection mode, but
		// later we will need to support persistent containers.
		select {
		case <-shutdownContext.Done():
			c.exec.kill()
		case <-c.exec.done():
		}
	}
}
