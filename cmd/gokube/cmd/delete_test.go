package cmd

import (
	"bytes"
	"errors"
	"testing"

	mocksdk "gokube/mocks/pkg/sdk"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDeleteCommand(t *testing.T) {
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

	t.Run("DeletePod", func(t *testing.T) {
		mockPodInterface.EXPECT().
			Delete(gomock.Any(), "test-pod").
			Return(nil).
			Times(1)

		cmd := NewDeleteCommand()
		cmd.SetArgs([]string{"pod", "test-pod"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `pod "test-pod" deleted`)
	})

	t.Run("DeletePodWithPlural", func(t *testing.T) {
		mockPodInterface.EXPECT().
			Delete(gomock.Any(), "test-pod2").
			Return(nil).
			Times(1)

		cmd := NewDeleteCommand()
		cmd.SetArgs([]string{"pods", "test-pod2"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `pod "test-pod2" deleted`)
	})

	t.Run("DeleteReplicaSet", func(t *testing.T) {
		mockReplicaSetInterface.EXPECT().
			Delete(gomock.Any(), "test-rs").
			Return(nil).
			Times(1)

		cmd := NewDeleteCommand()
		cmd.SetArgs([]string{"replicaset", "test-rs"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `replicaset "test-rs" deleted`)
	})

	t.Run("DeleteReplicaSetWithAlias", func(t *testing.T) {
		mockReplicaSetInterface.EXPECT().
			Delete(gomock.Any(), "test-rs2").
			Return(nil).
			Times(1)

		cmd := NewDeleteCommand()
		cmd.SetArgs([]string{"rs", "test-rs2"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `replicaset "test-rs2" deleted`)
	})

	t.Run("DeleteReplicaSetWithPlural", func(t *testing.T) {
		mockReplicaSetInterface.EXPECT().
			Delete(gomock.Any(), "test-rs3").
			Return(nil).
			Times(1)

		cmd := NewDeleteCommand()
		cmd.SetArgs([]string{"replicasets", "test-rs3"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, output.String(), `replicaset "test-rs3" deleted`)
	})

	t.Run("DeleteUnsupportedResource", func(t *testing.T) {
		cmd := NewDeleteCommand()
		cmd.SetArgs([]string{"invalid-resource", "test-name"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported resource type for delete")
	})

	t.Run("DeletePodError", func(t *testing.T) {
		mockPodInterface.EXPECT().
			Delete(gomock.Any(), "not-found").
			Return(errors.New("API request failed with status 404: Pod not found")).
			Times(1)

		cmd := NewDeleteCommand()
		cmd.SetArgs([]string{"pod", "not-found"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete pod")
	})

	t.Run("DeleteReplicaSetError", func(t *testing.T) {
		mockReplicaSetInterface.EXPECT().
			Delete(gomock.Any(), "not-found").
			Return(errors.New("API request failed with status 404: ReplicaSet not found")).
			Times(1)

		cmd := NewDeleteCommand()
		cmd.SetArgs([]string{"replicaset", "not-found"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete replicaset")
	})

	t.Run("DeleteMissingArguments", func(t *testing.T) {
		cmd := NewDeleteCommand()
		cmd.SetArgs([]string{"pod"})

		var output bytes.Buffer
		cmd.SetOut(&output)

		err := cmd.Execute()
		require.Error(t, err)
		// Command should require exactly 2 arguments
	})
}
