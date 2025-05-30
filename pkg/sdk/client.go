package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ClientInterface defines the interface for the gokube SDK client
//
//go:generate $PROJECT_HOME/bin/mock mocks/pkg/sdk
type ClientInterface interface {
	// Pods returns the pod interface
	Pods() PodInterface

	// Nodes returns the node interface
	Nodes() NodeInterface

	// ReplicaSets returns the replicaset interface
	ReplicaSets() ReplicaSetInterface
}

// Client represents the gokube SDK client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Config holds the configuration for the SDK client
type Config struct {
	BaseURL string
	Timeout time.Duration
}

// NewClient creates a new gokube SDK client
func NewClient(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// NewDefaultClient creates a new client with default configuration
func NewDefaultClient(baseURL string) *Client {
	return NewClient(Config{
		BaseURL: baseURL,
		Timeout: 30 * time.Second,
	})
}

// PodInterface provides methods for pod operations
func (c *Client) Pods() PodInterface {
	return &podClient{client: c}
}

// NodeInterface provides methods for node operations
func (c *Client) Nodes() NodeInterface {
	return &nodeClient{client: c}
}

// ReplicaSetInterface provides methods for replicaset operations
func (c *Client) ReplicaSets() ReplicaSetInterface {
	return &replicaSetClient{client: c}
}

// doRequest performs an HTTP request and handles the response
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}, params ...map[string]string) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	// Build URL with optional parameters
	var fullURL string
	if len(params) > 0 && params[0] != nil && len(params[0]) > 0 {
		fullURL = c.buildURL(path, params[0])
	} else {
		fullURL = c.baseURL + path
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// buildURL constructs a URL with optional query parameters
func (c *Client) buildURL(path string, params map[string]string) string {
	fullURL := c.baseURL + path

	if len(params) == 0 {
		return fullURL
	}

	u, err := url.Parse(fullURL)
	if err != nil {
		return fullURL
	}

	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	return u.String()
}
