package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gokube/pkg/api"
	"gokube/pkg/storage"
)

func TestCreateNode(t *testing.T) {
	etcdStorage, etcdServer := setupEtcdStorage()
	defer storage.StopEmbeddedEtcd(etcdServer)

	nodeRegistry := NewNodeRegistry(etcdStorage)
	node := createTestNode("test-node-1", "123")
	err := nodeRegistry.CreateNode(context.Background(), node)
	assert.NoError(t, err)
}

func TestGetNode(t *testing.T) {
	nodeName := "test-node-2"
	ctx := context.Background()
	etcdStorage, etcdServer := setupEtcdStorage()
	defer storage.StopEmbeddedEtcd(etcdServer)

	nodeRegistry := NewNodeRegistry(etcdStorage)

	createTestNodeInRegistry(t, nodeRegistry, nodeName, "456")

	node, err := nodeRegistry.GetNode(ctx, nodeName)
	assert.NoError(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, nodeName, node.Name)
	assert.Equal(t, "456", node.UID)
	assert.False(t, node.Spec.Unschedulable)
}

func TestGetNonExistentNode(t *testing.T) {
	ctx := context.Background()
	etcdStorage, etcdServer := setupEtcdStorage()
	defer storage.StopEmbeddedEtcd(etcdServer)

	nodeRegistry := NewNodeRegistry(etcdStorage)
	_, err := nodeRegistry.GetNode(ctx, "non-existent-node")
	assert.Error(t, err)
}

func TestUpdateNode(t *testing.T) {
	ctx := context.Background()
	etcdStorage, etcdServer := setupEtcdStorage()
	defer storage.StopEmbeddedEtcd(etcdServer)

	nodeRegistry := NewNodeRegistry(etcdStorage)
	nodeName := "test-node-3"
	createTestNodeInRegistry(t, nodeRegistry, nodeName, "789")

	node, err := nodeRegistry.GetNode(ctx, nodeName)
	require.NoError(t, err)

	node.Spec.Unschedulable = true
	err = nodeRegistry.UpdateNode(ctx, node)
	assert.NoError(t, err)

	updatedNode, err := nodeRegistry.GetNode(ctx, nodeName)
	assert.NoError(t, err)
	assert.True(t, updatedNode.Spec.Unschedulable)
}

func TestListNodes(t *testing.T) {
	ctx := context.Background()
	etcdStorage, etcdServer := setupEtcdStorage()
	defer storage.StopEmbeddedEtcd(etcdServer)

	nodeRegistry := NewNodeRegistry(etcdStorage)

	// Clear existing nodes
	clearNodes(t, nodeRegistry)

	// Create test nodes
	createTestNodeInRegistry(t, nodeRegistry, "test-node-4", "101")
	createTestNodeInRegistry(t, nodeRegistry, "test-node-5", "102")

	nodes, err := nodeRegistry.ListNodes(ctx)
	assert.NoError(t, err)
	assert.Len(t, nodes, 2)
	assert.Contains(t, []string{nodes[0].Name, nodes[1].Name}, "test-node-4")
	assert.Contains(t, []string{nodes[0].Name, nodes[1].Name}, "test-node-5")
}

func TestDeleteNode(t *testing.T) {
	ctx := context.Background()
	etcdStorage, etcdServer := setupEtcdStorage()
	defer storage.StopEmbeddedEtcd(etcdServer)

	nodeRegistry := NewNodeRegistry(etcdStorage)

	nodeName := "test-node-6"
	createTestNodeInRegistry(t, nodeRegistry, nodeName, "103")

	err := nodeRegistry.DeleteNode(ctx, nodeName)
	assert.NoError(t, err)

	_, err = nodeRegistry.GetNode(ctx, nodeName)
	assert.Error(t, err)
}

func TestDeleteNonExistentNode(t *testing.T) {
	ctx := context.Background()
	etcdStorage, etcdServer := setupEtcdStorage()
	defer storage.StopEmbeddedEtcd(etcdServer)

	nodeRegistry := NewNodeRegistry(etcdStorage)

	err := nodeRegistry.DeleteNode(ctx, "non-existent-node")
	assert.NoError(t, err) // Deleting a non-existent node should not return an error
}

// Helper functions

func createTestNode(name, uid string) *api.Node {
	return &api.Node{
		ObjectMeta: api.ObjectMeta{
			Name: name,
			UID:  uid,
		},
		Spec: api.NodeSpec{
			Unschedulable: false,
		},
	}
}

func createTestNodeInRegistry(t *testing.T, nodeRegistry *NodeRegistry, name, uid string) {
	node := createTestNode(name, uid)
	err := nodeRegistry.CreateNode(context.Background(), node)
	require.NoError(t, err)
}

func clearNodes(t *testing.T, nodeRegistry *NodeRegistry) {
	ctx := context.Background()
	nodes, err := nodeRegistry.ListNodes(ctx)
	require.NoError(t, err)
	for _, node := range nodes {
		err := nodeRegistry.DeleteNode(ctx, node.Name)
		require.NoError(t, err)
	}
}
