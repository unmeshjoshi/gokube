package registry

import (
	"context"
	"fmt"
	"path"

	"gokube/pkg/api"
	"gokube/pkg/storage"
)

const (
	nodePrefix = "/registry/nodes/"
)

// NodeRegistry provides CRUD operations for Node objects
type NodeRegistry struct {
	storage storage.Storage
}

// NewNodeRegistry creates a new NodeRegistry
func NewNodeRegistry(storage storage.Storage) *NodeRegistry {
	return &NodeRegistry{storage: storage}
}

// CreateNode stores a new Node
func (r *NodeRegistry) CreateNode(ctx context.Context, node *api.Node) error {
	key := path.Join(nodePrefix, node.Name)
	return r.storage.Create(ctx, key, node)
}

// GetNode retrieves a Node by name
func (r *NodeRegistry) GetNode(ctx context.Context, name string) (*api.Node, error) {
	key := path.Join(nodePrefix, name)
	node := &api.Node{}
	if err := r.storage.Get(ctx, key, node); err != nil {
		return nil, fmt.Errorf("error getting node: %v", err)
	}
	return node, nil
}

// UpdateNode updates an existing Node
func (r *NodeRegistry) UpdateNode(ctx context.Context, node *api.Node) error {
	key := path.Join(nodePrefix, node.Name)
	return r.storage.Update(ctx, key, node)
}

// DeleteNode removes a Node by name
func (r *NodeRegistry) DeleteNode(ctx context.Context, name string) error {
	key := path.Join(nodePrefix, name)
	return r.storage.Delete(ctx, key)
}

// ListNodes retrieves all Nodes
func (r *NodeRegistry) ListNodes(ctx context.Context) ([]*api.Node, error) {
	var nodes []*api.Node
	err := r.storage.List(ctx, nodePrefix, &nodes)
	if err != nil {
		return nil, fmt.Errorf("error listing nodes: %v", err)
	}
	return nodes, nil
}
