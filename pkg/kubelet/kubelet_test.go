package kubelet

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"gokube/pkg/api"
)

func TestStartContainerWithRealDocker(t *testing.T) {
	// Skip this test if we're not in an environment where we can connect to Docker
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		t.Skip("Skipping test: unable to connect to Docker")
	}

	ctx := context.Background()
	containerName := "test-container"
	imageName := "nginx"

	// Ensure the container doesn't exist before we start
	_ = dockerClient.ContainerRemove(ctx, containerName, container.RemoveOptions{Force: true})

	kubelet, err := NewKubelet("test-node", "http://fake-api-server-url")

	if err != nil {
		t.Fatalf("Failed to create Kubelet: %v", err)
	}

	pod := &api.Pod{
		ObjectMeta: api.ObjectMeta{Name: "test-pod"},
		NodeName:   "test-node",
		Spec: api.PodSpec{
			Containers: []api.Container{{Name: containerName, Image: imageName}},
		},
	}

	err = kubelet.runNewPods([]*api.Pod{pod})
	if err != nil {
		t.Fatalf("StartContainer failed: %v", err)
	}

	// Wait for the container to be created and running
	err = waitForContainer(ctx, dockerClient, containerName, 30*time.Second)
	if err != nil {
		t.Fatalf("Container did not start within the expected time: %v", err)
	}

	// Check if the container is running
	containerJSON, err := dockerClient.ContainerInspect(ctx, containerName)
	if err != nil {
		t.Fatalf("Failed to inspect container: %v", err)
	}

	if !containerJSON.State.Running {
		t.Errorf("Container is not running")
	}

	containerStatuses, err := kubelet.ListContainers(ctx)
	if err != nil {
		t.Errorf("Failed to list containers: %v", err)
	}

	fmt.Println(containerStatuses)

	if len(containerStatuses) != 1 {
		t.Errorf("Expected 1 container, got %d", len(containerStatuses))
	}

	// Clean up: stop and remove the container
	timeout := 10
	err = dockerClient.ContainerStop(ctx, containerName, container.StopOptions{Timeout: &timeout})
	if err != nil {
		t.Errorf("Failed to stop container: %v", err)
	}

	err = dockerClient.ContainerRemove(ctx, containerName, container.RemoveOptions{Force: true})
	if err != nil {
		t.Errorf("Failed to remove container: %v", err)
	}
}

func waitForContainer(ctx context.Context, client *client.Client, containerName string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for container to start")
		default:
			containerJSON, err := client.ContainerInspect(ctx, containerName)
			if err == nil && containerJSON.State.Running {
				return nil
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}
