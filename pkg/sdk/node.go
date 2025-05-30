package sdk

import (
	"context"
	"fmt"
	"net/http"

	"gokube/pkg/api"
)

// NodeInterface defines the interface for node operations
//
//go:generate $PROJECT_HOME/bin/mock mocks/pkg/sdk
type NodeInterface interface {
	// Get retrieves a node by name
	Get(ctx context.Context, name string) (*api.Node, error)

	// Update updates an existing node
	Update(ctx context.Context, node *api.Node) (*api.Node, error)

	// List retrieves all nodes
	List(ctx context.Context) ([]*api.Node, error)
}

// nodeClient implements NodeInterface
type nodeClient struct {
	client *Client
}

// Get retrieves a node by name
func (n *nodeClient) Get(ctx context.Context, name string) (*api.Node, error) {
	result := &api.Node{}
	path := fmt.Sprintf("/api/v1/nodes/%s", name)
	err := n.client.doRequest(ctx, http.MethodGet, path, nil, result)
	if err != nil {
		return nil, fmt.Errorf("failed to get node %s: %w", name, err)
	}
	return result, nil
}

// Update updates an existing node
func (n *nodeClient) Update(ctx context.Context, node *api.Node) (*api.Node, error) {
	result := &api.Node{}
	path := fmt.Sprintf("/api/v1/nodes/%s", node.Name)
	err := n.client.doRequest(ctx, http.MethodPut, path, node, result)
	if err != nil {
		return nil, fmt.Errorf("failed to update node %s: %w", node.Name, err)
	}
	return result, nil
}

// List retrieves all nodes
func (n *nodeClient) List(ctx context.Context) ([]*api.Node, error) {
	var result []*api.Node
	err := n.client.doRequest(ctx, http.MethodGet, "/api/v1/nodes", nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	if result == nil {
		result = []*api.Node{}
	}
	return result, nil
}
