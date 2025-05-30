package sdk

import (
	"context"
	"fmt"
	"net/http"

	"gokube/pkg/api"
)

// PodInterface defines the interface for pod operations
//
//go:generate $PROJECT_HOME/bin/mock mocks/pkg/sdk
type PodInterface interface {
	// Create creates a new pod
	Create(ctx context.Context, pod *api.Pod) (*api.Pod, error)

	// Get retrieves a pod by name
	Get(ctx context.Context, name string) (*api.Pod, error)

	// Update updates an existing pod
	Update(ctx context.Context, pod *api.Pod) (*api.Pod, error)

	// Delete deletes a pod by name
	Delete(ctx context.Context, name string) error

	// List retrieves all pods, optionally filtered by node name
	List(ctx context.Context, opts ...ListOption) ([]*api.Pod, error)

	// ListUnassigned retrieves all unassigned pods
	ListUnassigned(ctx context.Context) ([]*api.Pod, error)
}

// ListOptions holds options for listing resources
type ListOptions struct {
	NodeName string
}

// ListOption is a function that configures ListOptions
type ListOption func(*ListOptions)

// WithNodeName sets the node name filter for listing pods
func WithNodeName(nodeName string) ListOption {
	return func(opts *ListOptions) {
		opts.NodeName = nodeName
	}
}

// podClient implements PodInterface
type podClient struct {
	client *Client
}

// Create creates a new pod
func (p *podClient) Create(ctx context.Context, pod *api.Pod) (*api.Pod, error) {
	result := &api.Pod{}
	err := p.client.doRequest(ctx, http.MethodPost, "/api/v1/pods", pod, result)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %w", err)
	}
	return result, nil
}

// Get retrieves a pod by name
func (p *podClient) Get(ctx context.Context, name string) (*api.Pod, error) {
	result := &api.Pod{}
	path := fmt.Sprintf("/api/v1/pods/%s", name)
	err := p.client.doRequest(ctx, http.MethodGet, path, nil, result)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s: %w", name, err)
	}
	return result, nil
}

// Update updates an existing pod
func (p *podClient) Update(ctx context.Context, pod *api.Pod) (*api.Pod, error) {
	result := &api.Pod{}
	path := fmt.Sprintf("/api/v1/pods/%s", pod.Name)
	err := p.client.doRequest(ctx, http.MethodPut, path, pod, result)
	if err != nil {
		return nil, fmt.Errorf("failed to update pod %s: %w", pod.Name, err)
	}
	return result, nil
}

// Delete deletes a pod by name
func (p *podClient) Delete(ctx context.Context, name string) error {
	path := fmt.Sprintf("/api/v1/pods/%s", name)
	err := p.client.doRequest(ctx, http.MethodDelete, path, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete pod %s: %w", name, err)
	}
	return nil
}

// List retrieves all pods, optionally filtered by node name
func (p *podClient) List(ctx context.Context, opts ...ListOption) ([]*api.Pod, error) {
	options := &ListOptions{}
	for _, opt := range opts {
		opt(options)
	}

	params := make(map[string]string)
	if options.NodeName != "" {
		params["nodeName"] = options.NodeName
	}

	var result []*api.Pod
	path := "/api/v1/pods"

	err := p.client.doRequest(ctx, http.MethodGet, path, nil, &result, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if result == nil {
		result = []*api.Pod{}
	}
	return result, nil
}

// ListUnassigned retrieves all unassigned pods
func (p *podClient) ListUnassigned(ctx context.Context) ([]*api.Pod, error) {
	var result []*api.Pod
	err := p.client.doRequest(ctx, http.MethodGet, "/api/v1/pods/unassigned", nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to list unassigned pods: %w", err)
	}

	if result == nil {
		result = []*api.Pod{}
	}
	return result, nil
}
