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

func TestApplyCommand(t *testing.T) {
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

	t.Run("ApplyPodCreate", func(t *testing.T) {
		// Pod doesn't exist, should create
		mockPodInterface.EXPECT().
			Get(gomock.Any(), "new-pod").
			Return(nil, errors.New("API request failed with status 404: Pod not found")).
			Times(1)

		expectedPod := &api.Pod{
			ObjectMeta: api.ObjectMeta{Name: "new-pod"},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{Name: "main", Image: "nginx"},
				},
			},
		}

		mockPodInterface.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx interface{}, pod *api.Pod) (*api.Pod, error) {
				assert.Equal(t, "new-pod", pod.Name)
				assert.Equal(t, "nginx", pod.Spec.Containers[0].Image)
				return expectedPod, nil
			}).
			Times(1)

		cmd := NewApplyCommand()
		cmd.SetArgs([]string{"pod", "new-pod", "--image", "nginx"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `pod "new-pod" created`)
	})

	t.Run("ApplyPodUpdate", func(t *testing.T) {
		// Pod exists, should update
		existingPod := &api.Pod{
			ObjectMeta: api.ObjectMeta{Name: "existing-pod"},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{Name: "main", Image: "nginx"},
				},
			},
		}

		updatedPod := &api.Pod{
			ObjectMeta: api.ObjectMeta{Name: "existing-pod"},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{Name: "main", Image: "nginx:latest"},
				},
			},
		}

		mockPodInterface.EXPECT().
			Get(gomock.Any(), "existing-pod").
			Return(existingPod, nil).
			Times(1)

		mockPodInterface.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx interface{}, pod *api.Pod) (*api.Pod, error) {
				assert.Equal(t, "existing-pod", pod.Name)
				assert.Equal(t, "nginx:latest", pod.Spec.Containers[0].Image)
				return updatedPod, nil
			}).
			Times(1)

		cmd := NewApplyCommand()
		cmd.SetArgs([]string{"pod", "existing-pod", "--image", "nginx:latest"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `pod "existing-pod" configured`)
	})

	t.Run("ApplyReplicaSetCreate", func(t *testing.T) {
		// ReplicaSet doesn't exist, should create
		mockReplicaSetInterface.EXPECT().
			Get(gomock.Any(), "new-rs").
			Return(nil, errors.New("API request failed with status 404: ReplicaSet not found")).
			Times(1)

		expectedRS := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "new-rs"},
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
				assert.Equal(t, "new-rs", rs.Name)
				assert.Equal(t, int32(3), rs.Spec.Replicas)
				assert.Equal(t, "nginx", rs.Spec.Template.Spec.Containers[0].Image)
				return expectedRS, nil
			}).
			Times(1)

		cmd := NewApplyCommand()
		cmd.SetArgs([]string{"replicaset", "new-rs", "--replicas", "3", "--image", "nginx"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `replicaset "new-rs" created`)
	})

	t.Run("ApplyReplicaSetUpdate", func(t *testing.T) {
		// ReplicaSet exists, should update
		existingRS := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "existing-rs"},
			Spec: api.ReplicaSetSpec{
				Replicas: 2,
				Template: api.PodTemplateSpec{
					Spec: api.PodSpec{
						Containers: []api.Container{
							{Name: "main", Image: "nginx"},
						},
					},
				},
			},
		}

		updatedRS := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "existing-rs"},
			Spec: api.ReplicaSetSpec{
				Replicas: 5,
				Template: api.PodTemplateSpec{
					Spec: api.PodSpec{
						Containers: []api.Container{
							{Name: "main", Image: "nginx:latest"},
						},
					},
				},
			},
		}

		mockReplicaSetInterface.EXPECT().
			Get(gomock.Any(), "existing-rs").
			Return(existingRS, nil).
			Times(1)

		mockReplicaSetInterface.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx interface{}, rs *api.ReplicaSet) (*api.ReplicaSet, error) {
				assert.Equal(t, "existing-rs", rs.Name)
				assert.Equal(t, int32(5), rs.Spec.Replicas)
				assert.Equal(t, "nginx:latest", rs.Spec.Template.Spec.Containers[0].Image)
				return updatedRS, nil
			}).
			Times(1)

		cmd := NewApplyCommand()
		cmd.SetArgs([]string{"replicaset", "existing-rs", "--replicas", "5", "--image", "nginx:latest"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `replicaset "existing-rs" configured`)
	})

	t.Run("ApplyNodeUpdate", func(t *testing.T) {
		// Node exists, should update (nodes cannot be created)
		existingNode := &api.Node{
			ObjectMeta: api.ObjectMeta{Name: "worker-node"},
			Spec: api.NodeSpec{
				Unschedulable: false,
				ProviderID:    "aws://us-west-2/i-1234567890abcdef0",
			},
		}

		updatedNode := &api.Node{
			ObjectMeta: api.ObjectMeta{Name: "worker-node"},
			Spec: api.NodeSpec{
				Unschedulable: true,
				ProviderID:    "aws://us-west-2/i-1234567890abcdef0",
			},
		}

		mockNodeInterface.EXPECT().
			Get(gomock.Any(), "worker-node").
			Return(existingNode, nil).
			Times(1)

		mockNodeInterface.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx interface{}, node *api.Node) (*api.Node, error) {
				assert.Equal(t, "worker-node", node.Name)
				assert.Equal(t, true, node.Spec.Unschedulable)
				return updatedNode, nil
			}).
			Times(1)

		cmd := NewApplyCommand()
		cmd.SetArgs([]string{"node", "worker-node", "--unschedulable"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `node "worker-node" configured`)
	})

	t.Run("ApplyNodeNotFound", func(t *testing.T) {
		// Node doesn't exist and cannot be created
		mockNodeInterface.EXPECT().
			Get(gomock.Any(), "non-existent-node").
			Return(nil, errors.New("API request failed with status 404: Node not found")).
			Times(1)

		cmd := NewApplyCommand()
		cmd.SetArgs([]string{"node", "non-existent-node", "--unschedulable"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nodes cannot be created via CLI")
	})

	t.Run("ApplyUnsupportedResource", func(t *testing.T) {
		cmd := NewApplyCommand()
		cmd.SetArgs([]string{"invalid-resource", "test-name"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported resource type for apply")
	})

	t.Run("ApplyWithPlurals", func(t *testing.T) {
		// Test with plural resource names
		mockPodInterface.EXPECT().
			Get(gomock.Any(), "test-pod").
			Return(nil, errors.New("API request failed with status 404: Pod not found")).
			Times(1)

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
			Return(expectedPod, nil).
			Times(1)

		cmd := NewApplyCommand()
		cmd.SetArgs([]string{"pods", "test-pod"}) // using plural

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `pod "test-pod" created`)
	})

	t.Run("ApplyWithJSON", func(t *testing.T) {
		mockPodInterface.EXPECT().
			Get(gomock.Any(), "json-pod").
			Return(nil, errors.New("API request failed with status 404: Pod not found")).
			Times(1)

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

		cmd := NewApplyCommand()
		cmd.SetArgs([]string{"pod", "json-pod", "-o", "json"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `pod "json-pod" created`)
		assert.Contains(t, output.String(), `"name": "json-pod"`)
	})
}
