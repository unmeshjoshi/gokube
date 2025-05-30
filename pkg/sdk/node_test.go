package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gokube/pkg/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeOperations(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/nodes/test-node":
			// Get node
			node := &api.Node{
				ObjectMeta: api.ObjectMeta{Name: "test-node"},
				Spec: api.NodeSpec{
					Unschedulable: false,
					ProviderID:    "aws://us-west-2/i-1234567890abcdef0",
				},
				Status: api.NodeReady,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(node)

		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/nodes/test-node":
			// Update node
			var node api.Node
			err := json.NewDecoder(r.Body).Decode(&node)
			require.NoError(t, err)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(node)

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/nodes":
			// List nodes
			nodes := []*api.Node{
				{
					ObjectMeta: api.ObjectMeta{Name: "node1"},
					Spec: api.NodeSpec{
						Unschedulable: false,
						ProviderID:    "aws://us-west-2/i-1234567890abcdef1",
					},
					Status: api.NodeReady,
				},
				{
					ObjectMeta: api.ObjectMeta{Name: "node2"},
					Spec: api.NodeSpec{
						Unschedulable: true,
						ProviderID:    "aws://us-west-2/i-1234567890abcdef2",
					},
					Status: api.NodeNotReady,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(nodes)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewDefaultClient(server.URL)
	ctx := context.Background()

	t.Run("GetNode", func(t *testing.T) {
		node, err := client.Nodes().Get(ctx, "test-node")
		require.NoError(t, err)
		assert.Equal(t, "test-node", node.Name)
		assert.Equal(t, api.NodeReady, node.Status)
		assert.False(t, node.Spec.Unschedulable)
		assert.Equal(t, "aws://us-west-2/i-1234567890abcdef0", node.Spec.ProviderID)
	})

	t.Run("UpdateNode", func(t *testing.T) {
		node := &api.Node{
			ObjectMeta: api.ObjectMeta{Name: "test-node"},
			Spec: api.NodeSpec{
				Unschedulable: true,
				ProviderID:    "aws://us-west-2/i-updated",
			},
			Status: api.NodeNotReady,
		}

		result, err := client.Nodes().Update(ctx, node)
		require.NoError(t, err)
		assert.Equal(t, "test-node", result.Name)
		assert.True(t, result.Spec.Unschedulable)
		assert.Equal(t, api.NodeNotReady, result.Status)
		assert.Equal(t, "aws://us-west-2/i-updated", result.Spec.ProviderID)
	})

	t.Run("ListNodes", func(t *testing.T) {
		nodes, err := client.Nodes().List(ctx)
		require.NoError(t, err)
		assert.Len(t, nodes, 2)

		// Verify first node
		assert.Equal(t, "node1", nodes[0].Name)
		assert.Equal(t, api.NodeReady, nodes[0].Status)
		assert.False(t, nodes[0].Spec.Unschedulable)

		// Verify second node
		assert.Equal(t, "node2", nodes[1].Name)
		assert.Equal(t, api.NodeNotReady, nodes[1].Status)
		assert.True(t, nodes[1].Spec.Unschedulable)
	})
}

func TestNodeOperationsErrors(t *testing.T) {
	// Create test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/nodes/not-found":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Node not found"))
		case "/api/v1/nodes/server-error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
		case "/api/v1/nodes":
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to list nodes"))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewDefaultClient(server.URL)
	ctx := context.Background()

	t.Run("GetNodeNotFound", func(t *testing.T) {
		_, err := client.Nodes().Get(ctx, "not-found")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Contains(t, err.Error(), "failed to get node not-found")
	})

	t.Run("GetNodeServerError", func(t *testing.T) {
		_, err := client.Nodes().Get(ctx, "server-error")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("ListNodesError", func(t *testing.T) {
		_, err := client.Nodes().List(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
		assert.Contains(t, err.Error(), "failed to list nodes")
	})
}

func TestNodeInterface(t *testing.T) {
	client := NewDefaultClient("http://localhost:8080")

	// Verify that Nodes() returns the expected interface
	nodeInterface := client.Nodes()
	assert.NotNil(t, nodeInterface)

	// Type assertion to ensure it implements NodeInterface
	_, ok := nodeInterface.(NodeInterface)
	assert.True(t, ok, "Nodes() should return an implementation of NodeInterface")
}

func TestNodeStatusValues(t *testing.T) {
	// Test that we can use different node status values
	statuses := []api.NodeStatus{
		api.NodeReady,
		api.NodeNotReady,
		api.NodeMemoryPressure,
		api.NodeDiskPressure,
	}

	for _, status := range statuses {
		assert.NotEmpty(t, string(status), "Node status should not be empty")
	}
}
