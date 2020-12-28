# Changelog

## 0.9.1: JSON unmarshalling

The previous version of this library incorrectly unmarshalled JSON causing an endless loop. This release fixes JSON unmarshalling.

## 0.9.0: Split docker from dockerrun

In this release we split off the now-deprecated `dockerrun` backend. This new backend provides better compatibility and support for the [ContainerSSH Guest Agent](https://github.com/containerssh/agent).

Additionally, since we are switching to the Docker client v.20 the documentation will receive an addition explaining that the configuration server needs to be updated and deployed together with ContainerSSH.
