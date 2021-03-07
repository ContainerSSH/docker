package docker

// The ContainerSSH Docker module failed to read from the ContainerSSH agent. This
// is most likely because the ContainerSSH guest agent is not present in the guest image, but agent support is
// enabled.
const EFailedAgentRead = "DOCKER_AGENT_READ_FAILED"

// The ContainerSSH Docker module attempted to close the output (stdout and
// stderr) for writing but failed to do so.
const EFailedOutputCloseWriting = "DOCKER_CLOSE_OUTPUT_FAILED"

// The ContainerSSH Docker module attempted to close the input (stdin) for
// reading but failed to do so.
const EFailedInputCloseWriting = "DOCKER_CLOSE_INPUT_FAILED"

// The ContainerSSH Docker module detected a configuration error. Please check your
// configuration.
const EConfigError = "DOCKER_CONFIG_ERROR"

// The ContainerSSH Docker module is attaching to a container in session mode.
const MContainerAttach = "DOCKER_CONTAINER_ATTACH"

// The ContainerSSH Docker module has failed to attach to a container in
// session mode.
const EFailedContainerAttach = "DOCKER_CONTAINER_ATTACH_FAILED"

// The ContainerSSH Docker module is creating a container.
const MContainerCreate = "DOCKER_CONTAINER_CREATE"

// The ContainerSSH Docker module failed to create a container. This may be a
// temporary and retried or a permanent error message. Check the log message for details.
const EFailedContainerCreate = "DOCKER_CONTAINER_CREATE_FAILED"

// The ContainerSSH Docker module is starting the previously-created container.
const MContainerStart = "DOCKER_CONTAINER_START"

// The ContainerSSH docker module failed to start the container. This message
// can either be temporary and retried or permanent. Check the log message for details.
const EFailedContainerStart = "DOCKER_CONTAINER_START_FAILED"

// The ContainerSSH Docker module is stopping the container.
const MContainerStop = "DOCKER_CONTAINER_STOP"

// The ContainerSSH Docker module failed to stop the container. This message can
// be either temporary and retried or permanent. Check the log message for details.
const EContainerStopFailed = "DOCKER_CONTAINER_STOP_FAILED"

// The ContainerSSH Docker module os removing the container.
const MContainerRemove = "DOCKER_CONTAINER_REMOVE"

// The ContainerSSH Docker module could not remove the container. This message may be temporary and retried or
// permanent. Check the log message for details.
const EFailedContainerRemove = "DOCKER_CONTAINER_REMOVE_FAILED"

// The ContainerSSH Docker module has successfully removed the container.
const MContainerRemoveSuccessful = "DOCKER_CONTAINER_REMOVE_SUCCESSFUL"

// The ContainerSSH Docker module is sending a signal to the container.
const MContainerSignal = "DOCKER_CONTAINER_SIGNAL"

// The ContainerSSH Docker module has failed to send a signal to the
// container.
const EFailedContainerSignal = "DOCKER_CONTAINER_SIGNAL_FAILED"

// The ContainerSSH Docker module is shutting down a container.
const EShuttingDown = "DOCKER_CONTAINER_SHUTTING_DOWN"

// The ContainerSSH Docker module is creating an execution. This may be in connection mode, or
// it may be the module internally using the exec mechanism to deliver a payload into the container.
const MExec = "DOCKER_EXEC"

// The ContainerSSH Docker module is attaching to the previously-created execution.
const MExecAttach = "DOCKER_EXEC_ATTACH"

// The ContainerSSH Docker module could not attach to the previously-created
// execution.
const EFailedExecAttach = "DOCKER_EXEC_ATTACH_FAILED"

// The ContainerSSH Docker module is creating an execution.
const MExecCreate = "DOCKER_EXEC_CREATE"

// The ContainerSSH Docker module has failed to create an execution. This can be
// temporary and retried or permanent. See the error message for details.
const EFailedExecCreate = "DOCKER_EXEC_CREATE_FAILED"

// The ContainerSSH Docker module has failed to read the process ID from the
// [ContainerSSH Guest Agent](https://github.com/containerssh/agent). This is most likely because the guest image
// does not contain the guest agent, but guest agent support has been enabled.
const EFailedPIDRead = "DOCKER_EXEC_PID_READ_FAILED"

// The ContainerSSH Docker module is resizing the console.
const MResizing = "DOCKER_EXEC_RESIZE"

// The ContainerSSH Docker module failed to resize the console.
const EFailedResize = "DOCKER_EXEC_RESIZE_FAILED"

// The ContainerSSH Docker module is delivering a signal in container mode.
const MExecSignal = "DOCKER_EXEC_SIGNAL"

// The ContainerSSH Docker module failed to deliver a signal.
const EFailedExecSignal = "DOCKER_EXEC_SIGNAL_FAILED"

// The ContainerSSH Docker module failed to deliver a signal because
// [ContainerSSH Guest Agent](https://github.com/containerssh/agent) support is disabled.
const ECannotSendSignalNoAgent = "DOCKER_EXEC_SIGNAL_FAILED_NO_AGENT"

// The ContainerSSH Docker module successfully delivered the requested signal.
const MExecSignalSuccessful = "DOCKER_EXEC_SIGNAL_SUCCESSFUL"

// The ContainerSSH Docker module is fetching the exit code from the program.
const MExitCode = "DOCKER_EXIT_CODE"

// The ContainerSSH Docker module could not fetch the exit code from the program because the container is
// restarting. This is typically a misconfiguration as ContainerSSH containers should not automatically restart.
const EContainerRestarting = "DOCKER_EXIT_CODE_CONTAINER_RESTARTING"

// The ContainerSSH Docker module has failed to fetch the exit code of the
// program.
const EFetchingExitCodeFailed = "DOCKER_EXIT_CODE_FAILED"

// The ContainerSSH Docker module has received a negative exit code from Docker. This should never happen and is
// most likely a bug.
const ENegativeExitCode = "DOCKER_EXIT_CODE_NEGATIVE"

// The ContainerSSH Docker module could not fetch the program exit code because the
// program is still running. This error may be temporary and retried or permanent.
const EStillRunning = "DOCKER_EXIT_CODE_STILL_RUNNING"

// The ContainerSSH Docker module is listing the locally present container images to
// determine if the specified container image needs to be pulled.
const MImageList = "DOCKER_IMAGE_LISTING"

// The ContainerSSH Docker module failed to list the images present in the local Docker daemon. This is used to
// determine if the image needs to be pulled. This can be because the Docker daemon is not reachable, the
// certificate is invalid, or there is something else interfering with listing the images.
const EFailedImageList = "DOCKER_IMAGE_LISTING_FAILED"

// The ContainerSSH Docker module is pulling the container image.
const MImagePull = "DOCKER_IMAGE_PULL"

// The ContainerSSH Docker module failed to pull the specified container image. This can be because of connection
// issues to the Docker daemon, or because the Docker daemon itself can't pull the image. If you don't intend to
// have the image pulled you should set the `ImagePullPolicy` to `Never`. See the
// [Docker documentation](https://containerssh.io/reference/upcoming/docker) for details.
const EFailedImagePull = "DOCKER_IMAGE_PULL_FAILED"

// The ContainerSSH Docker module is checking if an image pull is needed.
const MImagePullNeeded = "DOCKER_IMAGE_PULL_NEEDED_CHECKING"

// The ContainerSSH Docker module can't execute the request because the
// program is already running. This is a client error.
const EProgramAlreadyRunning = "DOCKER_PROGRAM_ALREADY_RUNNING"

// The ContainerSSH Docker module can't deliver a signal because no PID has been
// recorded. This is most likely because guest agent support is disabled.
const EFailedSignalNoPID = "DOCKER_SIGNAL_FAILED_NO_PID"

// The ContainerSSH Docker module failed to stream stdin to the Docker engine.
const EFailedInputStream = "DOCKER_STREAM_INPUT_FAILED"

// The ContainerSSH Docker module failed to stream stdout and stderr from the
// Docker engine.
const EFailedOutputStream = "DOCKER_STREAM_OUTPUT_FAILED"

// The ContainerSSH Docker module is not configured to run the requested
// subsystem.
const ESubsystemNotSupported = "DOCKER_SUBSYSTEM_NOT_SUPPORTED"

// The [ContainerSSH Guest Agent](https://github.com/containerssh/agent) has been
// disabled, which is strongly discouraged. ContainerSSH requires the guest agent to be installed in the container
// image to facilitate all SSH features. Disabling the guest agent will result in breaking the expectations a user
// has towards an SSH server. We provide the ability to disable guest agent support only for cases where the guest
// agent binary cannot be installed in the image at all.
const EGuestAgentDisabled = "DOCKER_GUEST_AGENT_DISABLED"
