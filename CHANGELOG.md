# Changelog

## 0.9.3: Split docker from dockerrun

In this release we split the new backend using `docker exec` into a separate library because there were secuity-relevant breaking changes that we could not, in good conscience, fit into the `dockerrun` configuration.

The original `dockerrun` repository will be reverted to be compatible with the ContainerSSH 0.3 configuration structure. The `dockerrun` backend will receive a deprecation notice and be removed in ContainerSSH 0.5.

Additionally, since we are switching to the Docker client v.20 the documentation will receive an addition explaining that the configuration server needs to be updated and deployed together with ContainerSSH.

## 0.9.2: Bugfixing lock handling, bugfixing shell handling, security features

This release fixes several bugs related to lock handling on requests. It also fixes shell handling by introducing an explicit setting called `ShellCommand`. This setting must be explicitly configured to the shell that should be launched to avoid security problems with existing installations. Finally, we introduce two new switches called `DisableShell` and `DisableSubsystem` to disable shell and subsystem execution, respectively.

## 0.9.1: Retries

This release implements retrying API calls to the Docker API on failure, up until the value of the specified timeout.

## 0.9.0: Initial Release

This release ports the Dockerrun backend from [ContainerSSH 0.3](https://github.com/ContainerSSH/ContainerSSH). The previous version ran each session within a connection in a separate container. This release launches the container when the handshake is complete and then uses the `exec` functionality to launch programs for separate sessions.