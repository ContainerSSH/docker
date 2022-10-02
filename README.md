[![ContainerSSH - Launch Containers on Demand](https://containerssh.github.io/images/logo-for-embedding.svg)](https://containerssh.github.io/)

<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Docker Backend Library</h1>

<p align="center"><strong>⚠⚠⚠ Deprecated: ⚠⚠⚠</strong><br />This repository is deprecated in favor of <a href="https://github.com/ContainerSSH/libcontainerssh">libcontainerssh</a> for ContainerSSH 0.5.</p>

This library implements a backend that connects to a Docker socket and launches a new container for each connection, then runs executes a separate command per channel using `docker exec`. It replaces the legacy `dockerrun` backend.

## Using this library

This library implements a `NetworkConnectionHandler` from the [sshserver library](https://github.com/containerssh/sshserver). This can be embedded into a connection handler.

The network connection handler can be created with the `New()` method:

```go
var client net.TCPAddr
connectionID := "0123456789ABCDEF"
config := docker.Config{
    //...
}
collector := metrics.New()
dr, err := docker.New(
    client,
    connectionID,
    config,
    logger,
    collector.MustCreateCounter("backend_requests", "", ""),
    collector.MustCreateCounter("backend_failures", "", ""),
)
if err != nil {
    // Handle error
}
```

The `logger` parameter is a logger from the [ContainerSSH logger library](https://github.com/containerssh/log).

The `dr` variable can then be used to create a container on finished handshake:

```go
ssh, err := dr.OnHandshakeSuccess("provided-connection-username")
```

Conversely, on disconnect you must call `dr.OnDisconnect()`. The `ssh` variable can then be used to create session channels:

```go
var channelID uint64 = 0
extraData := []byte{}
session, err := ssh.OnSessionChannel(channelID, extraData)
```

Finally, the session can be used to launch programs:

```go
var requestID uint64 = 0
err = session.OnEnvRequest(requestID, "foo", "bar")
// ...
requestID = 1
var stdin io.Reader
var stdout, stderr io.Writer
err = session.OnShell(
    requestID,
    stdin,
    stdout,
    stderr,
    func(exitStatus ExitStatus) {
        // ...
    },
)
```

## Operating modes

This library supports several operating modes:

- `connection` creates a container per connection and uses the `docker exec` mechanism to launch SSH programs inside the container. This mode ignores the `CMD` of the container image and uses the `idleProgram` setting to launch inside the container.
- `session` creates a container per session and potentially results in multiple containers for a single SSH connection. This mode uses the `CMD` of the container image or from the configuration.
