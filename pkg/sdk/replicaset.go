package sdk

import (
	"context"
	"fmt"
	"net/http"

	"gokube/pkg/api"
)

// ReplicaSetInterface defines the interface for replicaset operations
//
//go:generate $PROJECT_HOME/bin/mock mocks/pkg/sdk
type ReplicaSetInterface interface {
	// Create creates a new replicaset
	Create(ctx context.Context, replicaset *api.ReplicaSet) (*api.ReplicaSet, error)

	// Get retrieves a replicaset by name
	Get(ctx context.Context, name string) (*api.ReplicaSet, error)

	// Update updates an existing replicaset
	Update(ctx context.Context, replicaset *api.ReplicaSet) (*api.ReplicaSet, error)

	// Delete deletes a replicaset by name
	Delete(ctx context.Context, name string) error

	// List retrieves all replicasets
	List(ctx context.Context) ([]*api.ReplicaSet, error)
}

// replicaSetClient implements ReplicaSetInterface
type replicaSetClient struct {
	client *Client
}

// Create creates a new replicaset
func (r *replicaSetClient) Create(ctx context.Context, replicaset *api.ReplicaSet) (*api.ReplicaSet, error) {
	result := &api.ReplicaSet{}
	err := r.client.doRequest(ctx, http.MethodPost, "/api/v1/replicasets", replicaset, result)
	if err != nil {
		return nil, fmt.Errorf("failed to create replicaset: %w", err)
	}
	return result, nil
}

// Get retrieves a replicaset by name
func (r *replicaSetClient) Get(ctx context.Context, name string) (*api.ReplicaSet, error) {
	result := &api.ReplicaSet{}
	path := fmt.Sprintf("/api/v1/replicasets/%s", name)
	err := r.client.doRequest(ctx, http.MethodGet, path, nil, result)
	if err != nil {
		return nil, fmt.Errorf("failed to get replicaset %s: %w", name, err)
	}
	return result, nil
}

// Update updates an existing replicaset
func (r *replicaSetClient) Update(ctx context.Context, replicaset *api.ReplicaSet) (*api.ReplicaSet, error) {
	result := &api.ReplicaSet{}
	path := fmt.Sprintf("/api/v1/replicasets/%s", replicaset.Name)
	err := r.client.doRequest(ctx, http.MethodPut, path, replicaset, result)
	if err != nil {
		return nil, fmt.Errorf("failed to update replicaset %s: %w", replicaset.Name, err)
	}
	return result, nil
}

// Delete deletes a replicaset by name
func (r *replicaSetClient) Delete(ctx context.Context, name string) error {
	path := fmt.Sprintf("/api/v1/replicasets/%s", name)
	err := r.client.doRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete replicaset %s: %w", name, err)
	}
	return nil
}

// List retrieves all replicasets
func (r *replicaSetClient) List(ctx context.Context) ([]*api.ReplicaSet, error) {
	var result []*api.ReplicaSet
	err := r.client.doRequest(ctx, http.MethodGet, "/api/v1/replicasets", nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to list replicasets: %w", err)
	}

	if result == nil {
		result = []*api.ReplicaSet{}
	}
	return result, nil
}
