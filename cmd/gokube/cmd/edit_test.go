package cmd

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	mocksdk "gokube/mocks/pkg/sdk"
	"gokube/pkg/api"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v2"
)

func TestNewEditCommand(t *testing.T) {
	cmd := NewEditCommand()

	assert.Equal(t, "edit", cmd.Name())
	assert.Equal(t, "Edit a resource in the default editor", cmd.Short)
	assert.True(t, strings.Contains(cmd.Long, "Edit a resource in your default editor"))
	assert.True(t, cmd.HasLocalFlags())
}

func TestNormalizeResourceType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"pod singular", "pod", "pod"},
		{"pod plural", "pods", "pod"},
		{"Pod uppercase", "Pod", "pod"},
		{"PODS uppercase", "PODS", "pod"},
		{"node singular", "node", "node"},
		{"node plural", "nodes", "node"},
		{"replicaset singular", "replicaset", "replicaset"},
		{"replicaset plural", "replicasets", "replicaset"},
		{"replicaset alias", "rs", "replicaset"},
		{"RS uppercase alias", "RS", "replicaset"},
		{"unknown resource", "unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeResourceType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRunEdit_UnsupportedResourceType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksdk.NewMockClientInterface(ctrl)
	SetTestClient(mockClient)
	defer ResetTestClient()

	cmd := NewEditCommand()
	args := []string{"unsupported", "test-name"}

	err := runEdit(cmd, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported resource type")
}

func TestRunEdit_PodNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksdk.NewMockClientInterface(ctrl)
	mockPods := mocksdk.NewMockPodInterface(ctrl)

	mockClient.EXPECT().Pods().Return(mockPods).AnyTimes()
	mockPods.EXPECT().Get(gomock.Any(), "nonexistent-pod").Return(nil, errors.New("pod not found"))

	SetTestClient(mockClient)
	defer ResetTestClient()

	cmd := NewEditCommand()
	args := []string{"pod", "nonexistent-pod"}

	err := runEdit(cmd, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get pod")
}

func TestRunEdit_NodeNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksdk.NewMockClientInterface(ctrl)
	mockNodes := mocksdk.NewMockNodeInterface(ctrl)

	mockClient.EXPECT().Nodes().Return(mockNodes).AnyTimes()
	mockNodes.EXPECT().Get(gomock.Any(), "nonexistent-node").Return(nil, errors.New("node not found"))

	SetTestClient(mockClient)
	defer ResetTestClient()

	cmd := NewEditCommand()
	args := []string{"node", "nonexistent-node"}

	err := runEdit(cmd, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get node")
}

func TestRunEdit_ReplicaSetNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksdk.NewMockClientInterface(ctrl)
	mockRS := mocksdk.NewMockReplicaSetInterface(ctrl)

	mockClient.EXPECT().ReplicaSets().Return(mockRS).AnyTimes()
	mockRS.EXPECT().Get(gomock.Any(), "nonexistent-rs").Return(nil, errors.New("replicaset not found"))

	SetTestClient(mockClient)
	defer ResetTestClient()

	cmd := NewEditCommand()
	args := []string{"replicaset", "nonexistent-rs"}

	err := runEdit(cmd, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get replicaset")
}

func TestGetEditor(t *testing.T) {
	tests := []struct {
		name     string
		editor   string
		visual   string
		expected string
	}{
		{"EDITOR set", "nano", "", "nano"},
		{"VISUAL set when EDITOR empty", "", "emacs", "emacs"},
		{"Both set, EDITOR takes precedence", "nano", "emacs", "nano"},
		{"Neither set, default to vim", "", "", "vim"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			originalEditor := os.Getenv("EDITOR")
			originalVisual := os.Getenv("VISUAL")

			// Set test values
			os.Setenv("EDITOR", tt.editor)
			os.Setenv("VISUAL", tt.visual)

			result := getEditor()
			assert.Equal(t, tt.expected, result)

			// Restore original values
			os.Setenv("EDITOR", originalEditor)
			os.Setenv("VISUAL", originalVisual)
		})
	}
}

func TestApplyEditedResource_InvalidYAML(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksdk.NewMockClientInterface(ctrl)
	SetTestClient(mockClient)
	defer ResetTestClient()

	ctx := context.Background()
	invalidYAML := []byte("invalid: yaml: content:")

	err := applyEditedResource(mockClient, ctx, "pod", "test-pod", invalidYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid YAML format")
}

func TestApplyEditedResource_NameMismatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksdk.NewMockClientInterface(ctrl)
	SetTestClient(mockClient)
	defer ResetTestClient()

	ctx := context.Background()

	pod := &api.Pod{
		ObjectMeta: api.ObjectMeta{Name: "wrong-name"},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{Name: "container1", Image: "nginx"},
			},
		},
	}

	podYAML, _ := yaml.Marshal(pod)

	err := applyEditedResource(mockClient, ctx, "pod", "expected-name", podYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource name cannot be changed")
}

func TestApplyEditedResource_PodUpdateSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksdk.NewMockClientInterface(ctrl)
	mockPods := mocksdk.NewMockPodInterface(ctrl)

	pod := &api.Pod{
		ObjectMeta: api.ObjectMeta{Name: "test-pod"},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{Name: "container1", Image: "nginx:latest"},
			},
		},
	}

	mockClient.EXPECT().Pods().Return(mockPods).AnyTimes()
	mockPods.EXPECT().Update(gomock.Any(), pod).Return(pod, nil)

	SetTestClient(mockClient)
	defer ResetTestClient()

	ctx := context.Background()
	podYAML, _ := yaml.Marshal(pod)

	err := applyEditedResource(mockClient, ctx, "pod", "test-pod", podYAML)
	assert.NoError(t, err)
}

func TestApplyEditedResource_NodeUpdateSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksdk.NewMockClientInterface(ctrl)
	mockNodes := mocksdk.NewMockNodeInterface(ctrl)

	node := &api.Node{
		ObjectMeta: api.ObjectMeta{Name: "test-node"},
		Spec: api.NodeSpec{
			Unschedulable: true,
		},
	}

	mockClient.EXPECT().Nodes().Return(mockNodes).AnyTimes()
	mockNodes.EXPECT().Update(gomock.Any(), node).Return(node, nil)

	SetTestClient(mockClient)
	defer ResetTestClient()

	ctx := context.Background()
	nodeYAML, _ := yaml.Marshal(node)

	err := applyEditedResource(mockClient, ctx, "node", "test-node", nodeYAML)
	assert.NoError(t, err)
}

func TestApplyEditedResource_ReplicaSetUpdateSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksdk.NewMockClientInterface(ctrl)
	mockRS := mocksdk.NewMockReplicaSetInterface(ctrl)

	rs := &api.ReplicaSet{
		ObjectMeta: api.ObjectMeta{Name: "test-rs"},
		Spec: api.ReplicaSetSpec{
			Replicas: 5,
		},
	}

	mockClient.EXPECT().ReplicaSets().Return(mockRS).AnyTimes()
	// Use gomock.Any() for the ReplicaSet parameter since YAML unmarshaling can change struct details
	// (e.g., populate default values in nested structures that weren't in the original)
	mockRS.EXPECT().Update(gomock.Any(), gomock.Any()).Return(rs, nil)

	SetTestClient(mockClient)
	defer ResetTestClient()

	ctx := context.Background()
	rsYAML, _ := yaml.Marshal(rs)

	err := applyEditedResource(mockClient, ctx, "replicaset", "test-rs", rsYAML)
	assert.NoError(t, err)
}

func TestApplyEditedResource_UpdateFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksdk.NewMockClientInterface(ctrl)
	mockPods := mocksdk.NewMockPodInterface(ctrl)

	pod := &api.Pod{
		ObjectMeta: api.ObjectMeta{Name: "test-pod"},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{Name: "container1", Image: "nginx"},
			},
		},
	}

	mockClient.EXPECT().Pods().Return(mockPods).AnyTimes()
	mockPods.EXPECT().Update(gomock.Any(), pod).Return(nil, errors.New("update failed"))

	SetTestClient(mockClient)
	defer ResetTestClient()

	ctx := context.Background()
	podYAML, _ := yaml.Marshal(pod)

	err := applyEditedResource(mockClient, ctx, "pod", "test-pod", podYAML)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update pod")
}

func TestApplyEditedResource_UnsupportedResourceType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocksdk.NewMockClientInterface(ctrl)
	SetTestClient(mockClient)
	defer ResetTestClient()

	ctx := context.Background()
	data := []byte("some: yaml")

	err := applyEditedResource(mockClient, ctx, "unsupported", "test-name", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported resource type")
}
