# GoKube: A Miniature Kubernetes-like Container Orchestrator
GoKube is an educational project that implements a simplified version of a container orchestrator, inspired by Kubernetes. This project is designed to teach the concepts of distributed system design using a Kubernetes-like system as an example.

## Project Overview

GoKube is built in Go and aims to demonstrate key concepts of container orchestration such as:

- Container scheduling
- Service discovery
- Load balancing
- State management
- Scaling

By implementing a miniature version of Kubernetes, this project provides hands-on experience with the fundamental principles of distributed systems and container orchestration.

## Prerequisites

- Basic understanding of Go programming language
- Familiarity with container concepts

## Setup
### 1. Homebrew
Install homebrew by following the instructions from [homebrew website](https://brew.sh/)

### 2. Docker
Install docker client using the following command:
```bash
brew install docker
```

### 3. Colima
This project recommends [colima](https://github.com/abiosoft/colima). Feel free to use alterantives like [Racher desktop](https://rancherdesktop.io/), [Docker desktop](https://www.docker.com/products/docker-desktop/),[Podman desktop](https://podman-desktop.io/),[Orbstack](https://orbstack.dev/) etc., if you are already familiar with it.

To install colima, run the following command from the project directory
```bash
make colima/install
```

Once colima is installed, run the following command to start colima immediatley and restart at login:
```bash
brew services start colima
```

Run the following command to start a VM:
```bash
make colima/start
```

Verify colima & docker is working by running the following command:
```bash
docker ps
```

Once it is verfied, add the following lines to your `~/.zshrc` or `~/.basrc` based on the type of your shell:

```bash
export DOCKER_HOST="unix://${HOME}/.colima/default/docker.sock"
```

### 4. Go
Install the latest version of golang from the official [website](https://go.dev/doc/install)

Verify golang is installed by running the following command:
```bash
go version
```

### 5. Process Compose
Install the latest version of [process-compose](https://f1bonacc1.github.io/process-compose/) which is required to run the project
```bash
brew install f1bonacc1/tap/process-compose
```

### 6. Xcode and command line tools

Ensure latest version of xcode and command line tools are installed

### 7. Devbox

Install the latest version of [devbox](https://www.jetify.com/docs/devbox/installing_devbox/) by running the following command:
```bash
curl -fsSL https://get.jetify.com/devbox | bash
```

### 8. Lima

Install the latest version of [lima](https://lima-vm.io/docs/) by running the following command:
```bash
brew install lima
```

### Running Commands

Run the following command to understand what make targets can be run:
```bash
make help
```
### Basic Commands

- To install dependencies - `make deps`
- To format code - `make fmt`
- To run vet - `make vet`
- To run lint - `make lint`
- To run all tests - `make test`
- To run package specific tests(api, controller, kubelet etc.,) - Eg: `make test/api`, `make test/controller`, `make test/kubelet`
- To generate mocks - `make mockgen`
- To build binaries - `make build`
- To build specific binaries - `make build/apiserver`, `make build/controller`, `make build/kubelet`
- To install binaries to GOPATH - `make install`
- To install specific binaries to GOPATH - `make install/apiserver`, `make install/controller`, `make install/kubelet`
- To run all necessary tasks before committing - `make precommit`
- To run the project - `make run`
- To clean the workspace - `make clean`

### Running the project
Run the following command to see the project in action:
```bash
make run
```

If there is a port conflict, you can change the port number variable `PORT` in `.env` file.

## Project Structure

The GoKube project is organized into several key directories:

```
gokube/
├── cmd/
│   ├── apiserver/
│   ├── controller/
│   ├── kubelet/
├── pkg/
│   ├── api/
│   ├── controller/
│   ├── kubelet/
│   ├── listwatch/
│   ├── registry/
│   ├── runtime/
│   ├── scheduler/
│   └── storage/
├── test/
│   └── ...
├── go.mod
├── go.sum
└── README.md

- `pkg/`: Contains the core packages used throughout the project.
  - `api/`: Defines the API objects and clients.
  - `controller/`: Implements the controllers for managing the system state.
  - `kubelet/`: Implements the kubelet functionality.
  - `listwatch/`: Implements the list and watch functionality.
  - `registry/`: Maintains the registry for k8s objects (nodes, pod, replicaset)
  - `runtime/`: Basic runtime utilities
  - `scheduler/`: Implements the scheduling of pods onto nodes.
  - `storage/`: Implements the storage handling via etcd

- `test/`: Contains integration and end-to-end tests.

This structure mimics Kubernetes' organization, providing a familiar layout for those acquainted with the Kubernetes codebase while simplifying it for educational purposes.

## Components

- API Server: Handles API requests and manages the system's state
- Kubelet: Manages containers on individual nodes
- Etcd: Distributed key-value store for system state (simulated)

## Current Features

- Basic container management (create, start, stop)
- Simple pod creation and management
- Rudimentary node management
- Robust etcd-based ListWatch implementation with:
  - Automatic reconnection with exponential backoff
  - Prometheus metrics for monitoring
  - Configurable retry behavior
  - Efficient event delivery via channels

## TODOs

The following features are planned for implementation:

1. [ ]Implement ReplicationController to create a specified number of replicas of created pods
2. [ ]Implement Scheduler to assign pods to nodes
3. [ ]Implement PodStatus update. Nodes should update pod status periodically with the apiserver
4. [ ]Update the ReplicationController to create newer instances of pods assigned to other nodes if a pod or node hosting the pods fails
5. [ ]Implement a Proxy service to load balance requests across pod instances~


## Learning Objectives

By working with this project, you will gain insights into:

1. The architecture of container orchestration systems
2. Distributed system design principles
3. Container lifecycle management
4. Network management in containerized environments
5. Challenges in distributed state management
6. Scaling and load balancing in distributed systems

## Acknowledgments

- Kubernetes project for inspiration
- Patterns Of Distributed Systems for design principles
```

## Running the application

To run the application, there are two more options that this project supports:
- `devbox`
- `limactl`

### 1: Using devbox

Once devbox is installed, navigate to the root directory of this project and run:

```bash
devbox shell
```

This will automatically install the required packages (`go`, `docker` and `colima`) and set up the environment. You can run the make commands from devbox shell.

```bash
devbox run app
```

### 2: Using limactl

If you prefer limactl use the following instructions:

After installing limactl, you can proceed with the rest of the setup instructions.

## Managing the VM

When the VM is started, it will have all the necessary tools installed, including Docker and etcd. Additionally, the path to the GoKube binary is set, allowing you to run the apiserver, controller, and kubelet directly from the VM shell.

The Makefile includes commands to manage a Lima VM for running GoKube. Here are the instructions to start, stop, delete, and access the VM shell.

### Initializing the VMs

To initialize the VMs, run the following command:

```bash
make lima/init-vms
```

### Start the VMs

To start the VMs, run:

```bash
make lima/start-vms
```

### Run the application

To start the processes in the master and worker VMs, run:

```bash
make lima/run
```

### Stop the VMs

To stop the VMs, run:

```bash
make lima/start-vms
```

### Cleanup

To cleanup use:

```bash
make lima/cleanup
```

This command will stop the lima instances and delete them

### Accessing the VM Shell

To access the shell of the running VM, execute:

```bash
make shell/master
make shell/worker1
make shell/worker2
```

This command will open a shell for the lima instance, allowing you to interact with the VM directly.