# Changelog

## 0.9.11: Better JSON and YAML support

Explicitly disabled internal fields in JSON or YAML.

## 0.9.10: Minor fixes

This release cleans up various log messages and adds small fixes.

## 0.9.9: Better logging, race condition fixes

This release adds extensive logging as well as fixes a few non-critical race conditions that lead to confusing error messages.

## 0.9.8: Conformance tests

This release makes use of the comprehensive conformance test introduced in [sshserver](https://github.com/containerssh/sshserver) and fixes a number of issues found in the process.

## 0.9.7: Regression bugfix

This release fixes a regression where non-TTY connections would not be handled correctly when running in `connection` mode.

## 0.9.6: Added Validate() for DockerRun

This release adds a `Validate()` method for the DockerRun backend.

## 0.9.5: Metrics integration

This release integrates the [metrics library](https://github.com/containerssh/metrics) and adds two parameters to `New` and `NewDockerRun` methods:

- `backendRequestsMetric` is a counter counting the number of requests to the Docker daemon.
- `backendFailuresMetric` is a counter counting the number of failed requests to the docker daemon.

## 0.9.4: Bug fix 

Fixed nil pointer on validation when host config is nil.

## 0.9.3: Nil support

In this release we are adding support for `nil` values when unmashalling JSON/YAML into time.Duration structures.

## 0.9.2: More compatibility fixes

This release adds more compatibility fixes with the 0.3 config format and the ability to unserialize durations from string instead of numbers.

## 0.9.1: JSON unmarshalling

The previous version of this library incorrectly unmarshalled JSON causing an endless loop. This release fixes JSON unmarshalling.

## 0.9.0: Split docker from dockerrun

In this release we split off the now-deprecated `dockerrun` backend. This new backend provides better compatibility and support for the [ContainerSSH Guest Agent](https://github.com/containerssh/agent).

Additionally, since we are switching to the Docker client v.20 the documentation will receive an addition explaining that the configuration server needs to be updated and deployed together with ContainerSSH.
