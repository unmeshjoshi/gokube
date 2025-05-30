package cmd

import (
	"bytes"
	"errors"
	"testing"

	mocksdk "gokube/mocks/pkg/sdk"
	"gokube/pkg/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock client and interfaces
	mockClient := mocksdk.NewMockClientInterface(ctrl)
	mockPodInterface := mocksdk.NewMockPodInterface(ctrl)
	mockReplicaSetInterface := mocksdk.NewMockReplicaSetInterface(ctrl)

	// Set up the mock client to return the mock interfaces
	mockClient.EXPECT().Pods().Return(mockPodInterface).AnyTimes()
	mockClient.EXPECT().ReplicaSets().Return(mockReplicaSetInterface).AnyTimes()

	// Set the test client
	SetTestClient(mockClient)
	defer ResetTestClient()

	t.Run("CreatePod", func(t *testing.T) {
		expectedPod := &api.Pod{
			ObjectMeta: api.ObjectMeta{Name: "test-pod"},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{Name: "main", Image: "nginx"},
				},
			},
		}

		mockPodInterface.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx interface{}, pod *api.Pod) (*api.Pod, error) {
				assert.Equal(t, "test-pod", pod.Name)
				assert.Equal(t, "nginx", pod.Spec.Containers[0].Image)
				return expectedPod, nil
			}).
			Times(1)

		cmd := NewCreateCommand()
		cmd.SetArgs([]string{"pod", "test-pod", "--image", "nginx"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), "test-pod")
	})

	t.Run("CreatePodWithCustomImage", func(t *testing.T) {
		expectedPod := &api.Pod{
			ObjectMeta: api.ObjectMeta{Name: "custom-pod"},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{Name: "main", Image: "alpine:latest"},
				},
			},
		}

		mockPodInterface.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx interface{}, pod *api.Pod) (*api.Pod, error) {
				assert.Equal(t, "custom-pod", pod.Name)
				assert.Equal(t, "alpine:latest", pod.Spec.Containers[0].Image)
				return expectedPod, nil
			}).
			Times(1)

		cmd := NewCreateCommand()
		cmd.SetArgs([]string{"pod", "custom-pod", "--image", "alpine:latest"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("CreateReplicaSet", func(t *testing.T) {
		expectedRS := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "test-rs"},
			Spec: api.ReplicaSetSpec{
				Replicas: 3,
				Template: api.PodTemplateSpec{
					Spec: api.PodSpec{
						Containers: []api.Container{
							{Name: "main", Image: "nginx"},
						},
					},
				},
			},
		}

		mockReplicaSetInterface.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx interface{}, rs *api.ReplicaSet) (*api.ReplicaSet, error) {
				assert.Equal(t, "test-rs", rs.Name)
				assert.Equal(t, int32(3), rs.Spec.Replicas)
				assert.Equal(t, "nginx", rs.Spec.Template.Spec.Containers[0].Image)
				return expectedRS, nil
			}).
			Times(1)

		cmd := NewCreateCommand()
		cmd.SetArgs([]string{"replicaset", "test-rs", "--replicas", "3", "--image", "nginx"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), "test-rs")
	})

	t.Run("CreateReplicaSetWithAlias", func(t *testing.T) {
		expectedRS := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "test-rs"},
			Spec: api.ReplicaSetSpec{
				Replicas: 2,
			},
		}

		mockReplicaSetInterface.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx interface{}, rs *api.ReplicaSet) (*api.ReplicaSet, error) {
				assert.Equal(t, "test-rs", rs.Name)
				assert.Equal(t, int32(2), rs.Spec.Replicas)
				return expectedRS, nil
			}).
			Times(1)

		cmd := NewCreateCommand()
		cmd.SetArgs([]string{"rs", "test-rs", "--replicas", "2"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("CreateUnsupportedResource", func(t *testing.T) {
		cmd := NewCreateCommand()
		cmd.SetArgs([]string{"invalid-resource", "test-name"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported resource type for create")
	})

	t.Run("CreatePodError", func(t *testing.T) {
		mockPodInterface.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("API request failed with status 409: Pod already exists")).
			Times(1)

		cmd := NewCreateCommand()
		cmd.SetArgs([]string{"pod", "existing-pod"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create pod")
	})

	t.Run("CreateWithJSON", func(t *testing.T) {
		expectedPod := &api.Pod{
			ObjectMeta: api.ObjectMeta{Name: "json-pod"},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{Name: "main", Image: "nginx"},
				},
			},
		}

		mockPodInterface.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(expectedPod, nil).
			Times(1)

		cmd := NewCreateCommand()
		cmd.SetArgs([]string{"pod", "json-pod", "-o", "json"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `"name": "json-pod"`)
	})
}
