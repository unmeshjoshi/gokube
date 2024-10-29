// File: pkg/kubelet/pod_status_test.go

package kubelet

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gokube/pkg/api"
)

func TestGetPodStatus(t *testing.T) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)
	defer dockerClient.Close()

	kubelet := &Kubelet{
		dockerClient: dockerClient,
	}

	tests := []struct {
		name            string
		containerNames  []string
		setupContainers func(t *testing.T, ctx context.Context, containerNames []string, dockerClient *client.Client) []string
		expectedStatus  api.PodStatus
	}{
		{
			name:           "All containers running",
			containerNames: []string{"c1", "c2", "c3"},
			setupContainers: func(t *testing.T, ctx context.Context, containerNames []string, dockerClient *client.Client) []string {
				return createContainers(t, ctx, dockerClient, containerNames, func(config *container.Config) {
					config.Cmd = []string{"sleep", "infinity"}
				})
			},
			expectedStatus: api.PodRunning,
		},
		{
			name:           "One container running, others completed successfully",
			containerNames: []string{"c1", "c2", "c3"},
			setupContainers: func(t *testing.T, ctx context.Context, containerNames []string, dockerClient *client.Client) []string {
				existedIds := createContainers(t, ctx, dockerClient, []string{containerNames[0], containerNames[1]}, func(config *container.Config) {
					config.Cmd = []string{"echo", "success"}
				})
				// Keep one container running
				runningContainerIds := createContainers(t, ctx, dockerClient, []string{containerNames[2]}, func(config *container.Config) {
					config.Cmd = []string{"sleep", "infinity"}
				})
				require.NoError(t, err)
				return append(existedIds, runningContainerIds...)
			},
			expectedStatus: api.PodRunning,
		},
		{
			name:           "All containers completed successfully",
			containerNames: []string{"c1", "c2", "c3"},
			setupContainers: func(t *testing.T, ctx context.Context, containerNames []string, dockerClient *client.Client) []string {
				return createContainers(t, ctx, dockerClient, containerNames, func(config *container.Config) {
					config.Cmd = []string{"echo", "success"}
				})
			},
			expectedStatus: api.PodSucceeded,
		},
		{
			name:           "All containers failed",
			containerNames: []string{"c1", "c2", "c3"},
			setupContainers: func(t *testing.T, ctx context.Context, containerNames []string, dockerClient *client.Client) []string {
				return createContainers(t, ctx, dockerClient, containerNames, func(config *container.Config) {
					config.Cmd = []string{"sh", "-c", "exit 1"}
				})
			},
			expectedStatus: api.PodFailed,
		},
		{
			name:           "Mixed container states",
			containerNames: []string{"c1", "c2", "c3"},
			setupContainers: func(t *testing.T, ctx context.Context, containerNames []string, dockerClient *client.Client) []string {
				ids := make([]string, 3)

				// Running container
				runningIDs := createContainers(t, ctx, dockerClient, []string{containerNames[0]}, func(config *container.Config) {
					config.Cmd = []string{"sleep", "infinity"}
				})
				ids[0] = runningIDs[0]

				// Completed container
				completedIDs := createContainers(t, ctx, dockerClient, []string{containerNames[1]}, func(config *container.Config) {
					config.Cmd = []string{"echo", "success"}
				})
				ids[1] = completedIDs[0]

				// Failed container
				failedIDs := createContainers(t, ctx, dockerClient, []string{containerNames[2]}, func(config *container.Config) {
					config.Cmd = []string{"sh", "-c", "exit 1"}
				})
				ids[2] = failedIDs[0]

				return ids
			},
			expectedStatus: api.PodRunning,
		},
		{
			name:           "No containers created",
			containerNames: []string{},
			setupContainers: func(t *testing.T, ctx context.Context, containerNames []string, dockerClient *client.Client) []string {
				return []string{}
			},
			expectedStatus: api.PodSucceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			containerIDs := tt.setupContainers(t, ctx, tt.containerNames, dockerClient)
			defer removeContainers(t, ctx, dockerClient, containerIDs)

			pod := &api.Pod{
				ObjectMeta: api.ObjectMeta{Name: "test-pod"},
				Spec: api.PodSpec{
					Containers: make([]api.Container, len(tt.containerNames)),
				},
			}
			for i := range tt.containerNames {
				pod.Spec.Containers[i] = api.Container{Name: tt.containerNames[i]}
			}

			status, err := kubelet.getPodStatus(ctx, pod)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func createContainers(t *testing.T, ctx context.Context, dockerClient *client.Client, containerNames []string, configModifier func(*container.Config)) []string {
	ids := make([]string, len(containerNames))
	for i, name := range containerNames {
		config := &container.Config{
			Image: "alpine:latest",
		}
		configModifier(config)

		resp, err := dockerClient.ContainerCreate(ctx, config, nil, nil, nil, name)
		require.NoError(t, err)
		ids[i] = resp.ID

		err = dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{})
		require.NoError(t, err)
	}

	// Give containers time to run
	time.Sleep(2 * time.Second)
	return ids
}

func removeContainers(t *testing.T, ctx context.Context, dockerClient *client.Client, containerIDs []string) {
	for _, id := range containerIDs {
		err := dockerClient.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
		assert.NoError(t, err)
	}
}
