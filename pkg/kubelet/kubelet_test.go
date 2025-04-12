package kubelet

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
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
	podName := "test-pod"
	containerName := "test-container"
	imageName := "nginx"
	uniqueContainerName := fmt.Sprintf("%s-%s", podName, containerName)
	containerIds := listContainerIDs(ctx, dockerClient, uniqueContainerName)

	// Ensure the container doesn't exist before we start
	for _, containerId := range containerIds {
		_ = dockerClient.ContainerRemove(ctx, containerId, container.RemoveOptions{Force: true})
	}

	kubelet, err := NewKubelet("test-node", "http://fake-api-server-url")

	if err != nil {
		t.Fatalf("Failed to create Kubelet: %v", err)
	}

	pod := &api.Pod{
		ObjectMeta: api.ObjectMeta{Name: podName},
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
	containerId, err := waitForContainer(ctx, dockerClient, uniqueContainerName, 60*time.Second)
	if err != nil {
		t.Fatalf("Container did not start within the expected time: %v", err)
	}

	// Check if the container is running
	containerJSON, err := dockerClient.ContainerInspect(ctx, containerId)
	if err != nil {
		fmt.Printf("Failed to inspect container: %v\n", err)
	}

	fmt.Printf("Container state: %+v\n", containerJSON.State)

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
	err = dockerClient.ContainerStop(ctx, containerId, container.StopOptions{Timeout: &timeout})
	if err != nil {
		t.Errorf("Failed to stop container: %v", err)
	}

	err = dockerClient.ContainerRemove(ctx, containerId, container.RemoveOptions{Force: true})
	if err != nil {
		t.Errorf("Failed to remove container: %v", err)
	}
}

func waitForContainer(ctx context.Context, client *client.Client, containerName string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("timeout waiting for container to start")
		default:
			containerIds := listContainerIDs(ctx, client, containerName)
			if len(containerIds) > 0 {
				containerId := containerIds[1]
				containerJSON, err := client.ContainerInspect(ctx, containerId)
				if err != nil {
					fmt.Printf("Failed to inspect container %v: %v\n", containerId, err)
				}
				if err == nil && containerJSON.State.Running {
					fmt.Printf("Container state: %+v\n", containerJSON.State)
					return containerId, nil
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func listContainerIDs(ctx context.Context, dockerClient *client.Client, dockerContainerName string) []string {
	listFilters := filters.NewArgs(filters.Arg("name", dockerContainerName))
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{Filters: listFilters})
	if err != nil {
		fmt.Printf("Failed to list containers %v\n", err)
	}
	containerIds := make([]string, len(containers))
	for _, container := range containers {
		if len(container.ID) > 0 {
			containerIds = append(containerIds, container.ID)
		}
	}
	return containerIds
}
