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
- To build specific binaries - `make build/apiserver`, `make build/controller`, `make build/kubelet`, `make build/scheduler`, `make build/gokube`
- To install binaries to GOPATH - `make install`
- To install specific binaries to GOPATH - `make install/apiserver`, `make install/controller`, `make install/kubelet`, `make install/scheduler`, `make install/gokube`
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
│   ├── scheduler/
│   └── gokube/           # CLI tool for managing gokube resources
├── pkg/
│   ├── api/
│   ├── controller/
│   ├── kubelet/
│   ├── listwatch/
│   ├── registry/
│   ├── runtime/
│   ├── scheduler/
│   ├── sdk/              # GoKube SDK for client operations
│   └── storage/
├── test/
│   └── ...
├── go.mod
├── go.sum
└── README.md

- `cmd/`: Contains the main applications and CLI tools.
  - `apiserver/`: The GoKube API server implementation.
  - `controller/`: The controller manager implementation.
  - `kubelet/`: The kubelet implementation for node management.
  - `scheduler/`: The scheduler for pod placement.
  - `gokube/`: A kubectl-like CLI tool for managing GoKube resources.

- `pkg/`: Contains the core packages used throughout the project.
  - `api/`: Defines the API objects and clients.
  - `controller/`: Implements the controllers for managing the system state.
  - `kubelet/`: Implements the kubelet functionality.
  - `listwatch/`: Implements the list and watch functionality.
  - `registry/`: Maintains the registry for k8s objects (nodes, pod, replicaset)
  - `runtime/`: Basic runtime utilities
  - `scheduler/`: Implements the scheduling of pods onto nodes.
  - `sdk/`: Go SDK for interacting with GoKube APIs programmatically.
  - `storage/`: Implements the storage handling via etcd

- `test/`: Contains integration and end-to-end tests.

This structure mimics Kubernetes' organization, providing a familiar layout for those acquainted with the Kubernetes codebase while simplifying it for educational purposes.

## Components

- **API Server**: Handles API requests and manages the system's state
- **Controller**: Manages the desired state of resources like ReplicaSets
- **Kubelet**: Manages containers on individual nodes
- **Scheduler**: Assigns pods to appropriate nodes based on resource requirements
- **CLI Tool (gokube)**: A kubectl-like command-line interface for resource management
- **SDK**: A Go client library for programmatic access to GoKube APIs
- **Etcd**: Distributed key-value store for system state (simulated)

## GoKube CLI Tool

The project now includes a comprehensive CLI tool called `gokube` that provides kubectl-like functionality for managing GoKube resources.

### CLI Features

- **Resource Management**: Create, read, update, and delete pods, nodes, and replicasets
- **Multiple Output Formats**: Support for table, JSON, and YAML output formats
- **Resource Filtering**: Filter pods by node assignment or show unassigned pods
- **Scaling Operations**: Scale replicasets up or down
- **Interactive Editing**: Edit resources in your preferred editor

### CLI Commands

```bash
# Build and install the CLI
make build/gokube
make install/gokube

# Get resources
gokube get pods
gokube get pod my-pod
gokube get nodes
gokube get replicasets

# Create resources
gokube create pod my-pod --image nginx
gokube create replicaset my-rs --replicas 3 --image nginx

# Apply resources (create or update)
gokube apply pod my-pod --image nginx:latest
gokube apply replicaset my-rs --replicas 5

# Edit resources interactively
gokube edit pod my-pod
gokube edit replicaset my-rs

# Delete resources
gokube delete pod my-pod
gokube delete replicaset my-rs

# Scale replicasets
gokube scale replicaset my-rs --replicas 10

# Get help
gokube --help
gokube get --help
```

### Output Formats

The CLI supports multiple output formats:

```bash
# Table format (default)
gokube get pods

# JSON format
gokube get pods -o json

# YAML format
gokube get pods -o yaml
```

### Filtering Options

```bash
# Get pods on a specific node
gokube get pods --node worker-1

# Get unassigned pods
gokube get pods --unassigned
```

## GoKube SDK

The project includes a comprehensive Go SDK (`pkg/sdk`) that provides programmatic access to GoKube APIs.

### SDK Features

- **Type-Safe Client**: Strongly-typed Go client for all GoKube resources
- **Context Support**: Full context.Context support for timeouts and cancellation
- **Configurable**: Customizable base URL, timeouts, and HTTP client settings
- **Resource Interfaces**: Separate interfaces for pods, nodes, and replicasets
- **Testing Support**: Mock-friendly interfaces for unit testing

### SDK Usage

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "gokube/pkg/sdk"
    "gokube/pkg/api"
)

func main() {
    // Create a client
    config := sdk.Config{
        BaseURL: "http://localhost:8080",
        Timeout: 30 * time.Second,
    }
    client := sdk.NewClient(config)
    
    // Or use default configuration
    client := sdk.NewDefaultClient("http://localhost:8080")
    
    ctx := context.Background()
    
    // Work with pods
    pods, err := client.Pods().List(ctx)
    if err != nil {
        panic(err)
    }
    
    pod, err := client.Pods().Get(ctx, "my-pod")
    if err != nil {
        panic(err)
    }
    
    newPod := &api.Pod{
        Metadata: api.Metadata{
            Name: "test-pod",
        },
        Spec: api.PodSpec{
            Image: "nginx:latest",
        },
    }
    
    created, err := client.Pods().Create(ctx, newPod)
    if err != nil {
        panic(err)
    }
    
    // Work with nodes
    nodes, err := client.Nodes().List(ctx)
    if err != nil {
        panic(err)
    }
    
    // Work with replicasets
    replicasets, err := client.ReplicaSets().List(ctx)
    if err != nil {
        panic(err)
    }
}
```

### SDK Interfaces

The SDK provides the following main interfaces:

- **`ClientInterface`**: Main client interface providing access to resource clients
- **`PodInterface`**: Operations for pod management (CRUD, list by node, list unassigned)
- **`NodeInterface`**: Operations for node management
- **`ReplicaSetInterface`**: Operations for replicaset management

## Current Features

- **Container Management**: Basic container operations including create, start, and stop functionality
- **Pod Management**: Simple pod creation and lifecycle management
- **Node Management**: Basic node registration and management capabilities
- **ReplicaSet Management**: Maintains desired pod replicas with automatic scaling and reconciliation
- **Pod Scheduling**: Assigns pods to available nodes based on resource requirements
- **ReplicaSet Controller**: Automatically maintains desired number of pod replicas with scaling, reconciliation, and failure recovery
- **Scheduler**: Assigns pending pods to available nodes with configurable scheduling intervals
- **Pod Status Updates**: Kubelet monitors and updates pod status periodically (every 10 seconds) based on container states
- **Failure Handling**: ReplicaSet controller recreates failed pods on healthy nodes through reconciliation loops
- **Robust ListWatch Implementation**: etcd-based implementation with:
  - Automatic reconnection with exponential backoff
  - Prometheus metrics for monitoring
  - Configurable retry behavior
  - Efficient event delivery via channels
- **GoKube CLI Tool**: kubectl-like interface for resource management
- **GoKube SDK**: Type-safe Go client library for programmatic API access
- **Multiple Output Formats**: Table, JSON, and YAML support in CLI
- **Resource Filtering**: Advanced filtering options for resource queries

## TODOs

The following features are planned for implementation:

1. [ ] Implement a Proxy service to load balance requests across pod instances (kube-proxy equivalent with iptables rules)

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