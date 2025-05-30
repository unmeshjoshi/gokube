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

func TestScaleCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock client and interfaces
	mockClient := mocksdk.NewMockClientInterface(ctrl)
	mockReplicaSetInterface := mocksdk.NewMockReplicaSetInterface(ctrl)

	// Set up the mock client to return the mock interfaces
	mockClient.EXPECT().ReplicaSets().Return(mockReplicaSetInterface).AnyTimes()

	// Set the test client
	SetTestClient(mockClient)
	defer ResetTestClient()

	t.Run("ScaleReplicaSet", func(t *testing.T) {
		existingRS := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "test-rs"},
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
			Status: api.ReplicaSetStatus{
				Replicas:      2,
				ReadyReplicas: 2,
			},
		}

		scaledRS := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "test-rs"},
			Spec: api.ReplicaSetSpec{
				Replicas: 5,
				Template: api.PodTemplateSpec{
					Spec: api.PodSpec{
						Containers: []api.Container{
							{Name: "main", Image: "nginx"},
						},
					},
				},
			},
			Status: api.ReplicaSetStatus{
				Replicas:      5,
				ReadyReplicas: 5,
			},
		}

		mockReplicaSetInterface.EXPECT().
			Get(gomock.Any(), "test-rs").
			Return(existingRS, nil).
			Times(1)

		mockReplicaSetInterface.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx interface{}, rs *api.ReplicaSet) (*api.ReplicaSet, error) {
				assert.Equal(t, "test-rs", rs.Name)
				assert.Equal(t, int32(5), rs.Spec.Replicas)
				return scaledRS, nil
			}).
			Times(1)

		cmd := NewScaleCommand()
		cmd.SetArgs([]string{"replicaset", "test-rs", "--replicas", "5"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `replicaset "test-rs" scaled to 5 replicas`)
		assert.Contains(t, output.String(), "test-rs")
	})

	t.Run("ScaleReplicaSetWithAlias", func(t *testing.T) {
		existingRS := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "test-rs"},
			Spec: api.ReplicaSetSpec{
				Replicas: 3,
			},
		}

		scaledRS := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "test-rs"},
			Spec: api.ReplicaSetSpec{
				Replicas: 1,
			},
		}

		mockReplicaSetInterface.EXPECT().
			Get(gomock.Any(), "test-rs").
			Return(existingRS, nil).
			Times(1)

		mockReplicaSetInterface.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx interface{}, rs *api.ReplicaSet) (*api.ReplicaSet, error) {
				assert.Equal(t, int32(1), rs.Spec.Replicas)
				return scaledRS, nil
			}).
			Times(1)

		cmd := NewScaleCommand()
		cmd.SetArgs([]string{"rs", "test-rs", "--replicas", "1"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `replicaset "test-rs" scaled to 1 replicas`)
	})

	t.Run("ScaleReplicaSetWithPlural", func(t *testing.T) {
		existingRS := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "test-rs"},
			Spec: api.ReplicaSetSpec{
				Replicas: 2,
			},
		}

		scaledRS := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "test-rs"},
			Spec: api.ReplicaSetSpec{
				Replicas: 10,
			},
		}

		mockReplicaSetInterface.EXPECT().
			Get(gomock.Any(), "test-rs").
			Return(existingRS, nil).
			Times(1)

		mockReplicaSetInterface.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(scaledRS, nil).
			Times(1)

		cmd := NewScaleCommand()
		cmd.SetArgs([]string{"replicasets", "test-rs", "--replicas", "10"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `replicaset "test-rs" scaled to 10 replicas`)
	})

	t.Run("ScaleUnsupportedResource", func(t *testing.T) {
		cmd := NewScaleCommand()
		cmd.SetArgs([]string{"pod", "test-pod", "--replicas", "5"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported resource type for scale")
	})

	t.Run("ScaleReplicaSetNotFound", func(t *testing.T) {
		mockReplicaSetInterface.EXPECT().
			Get(gomock.Any(), "not-found").
			Return(nil, errors.New("API request failed with status 404: ReplicaSet not found")).
			Times(1)

		cmd := NewScaleCommand()
		cmd.SetArgs([]string{"replicaset", "not-found", "--replicas", "3"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get replicaset")
	})

	t.Run("ScaleReplicaSetUpdateError", func(t *testing.T) {
		existingRS := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "test-rs"},
			Spec: api.ReplicaSetSpec{
				Replicas: 2,
			},
		}

		mockReplicaSetInterface.EXPECT().
			Get(gomock.Any(), "test-rs").
			Return(existingRS, nil).
			Times(1)

		mockReplicaSetInterface.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("API request failed with status 409: Conflict")).
			Times(1)

		cmd := NewScaleCommand()
		cmd.SetArgs([]string{"replicaset", "test-rs", "--replicas", "5"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scale replicaset")
	})

	t.Run("ScaleWithJSON", func(t *testing.T) {
		existingRS := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "test-rs"},
			Spec: api.ReplicaSetSpec{
				Replicas: 2,
			},
		}

		scaledRS := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "test-rs"},
			Spec: api.ReplicaSetSpec{
				Replicas: 3,
			},
		}

		mockReplicaSetInterface.EXPECT().
			Get(gomock.Any(), "test-rs").
			Return(existingRS, nil).
			Times(1)

		mockReplicaSetInterface.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(scaledRS, nil).
			Times(1)

		cmd := NewScaleCommand()
		cmd.SetArgs([]string{"replicaset", "test-rs", "--replicas", "3", "-o", "json"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `replicaset "test-rs" scaled to 3 replicas`)
		assert.Contains(t, output.String(), `"name": "test-rs"`)
	})

	t.Run("ScaleMissingReplicasFlag", func(t *testing.T) {
		cmd := NewScaleCommand()
		cmd.SetArgs([]string{"replicaset", "test-rs"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.Error(t, err)
		// Should fail because --replicas flag is required
	})
}
