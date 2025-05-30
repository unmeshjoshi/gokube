package cmd

import (
	"bytes"
	"testing"

	mocksdk "gokube/mocks/pkg/sdk"
	"gokube/pkg/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock client and interfaces
	mockClient := mocksdk.NewMockClientInterface(ctrl)
	mockPodInterface := mocksdk.NewMockPodInterface(ctrl)
	mockNodeInterface := mocksdk.NewMockNodeInterface(ctrl)
	mockReplicaSetInterface := mocksdk.NewMockReplicaSetInterface(ctrl)

	// Set up the mock client to return the mock interfaces
	mockClient.EXPECT().Pods().Return(mockPodInterface).AnyTimes()
	mockClient.EXPECT().Nodes().Return(mockNodeInterface).AnyTimes()
	mockClient.EXPECT().ReplicaSets().Return(mockReplicaSetInterface).AnyTimes()

	// Set the test client
	SetTestClient(mockClient)
	defer ResetTestClient()

	t.Run("GetPod", func(t *testing.T) {
		expectedPod := &api.Pod{
			ObjectMeta: api.ObjectMeta{Name: "test-pod"},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{Name: "main", Image: "nginx"},
				},
			},
			Status:   api.PodRunning,
			NodeName: "node1",
		}

		mockPodInterface.EXPECT().
			Get(gomock.Any(), "test-pod").
			Return(expectedPod, nil).
			Times(1)

		cmd := NewGetCommand()
		cmd.SetArgs([]string{"pod", "test-pod"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), "test-pod")
		assert.Contains(t, output.String(), "Running")
	})

	t.Run("GetUnsupportedResource", func(t *testing.T) {
		cmd := NewGetCommand()
		cmd.SetArgs([]string{"invalid-resource"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported resource type")
	})
}
