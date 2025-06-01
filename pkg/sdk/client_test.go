package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	config := Config{
		BaseURL: "http://localhost:8080",
		Timeout: 10 * time.Second,
	}

	client := NewClient(config)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8080", client.baseURL)
	assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
}

func TestNewDefaultClient(t *testing.T) {
	client := NewDefaultClient("http://localhost:8080")
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8080", client.baseURL)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestClientInterfaces(t *testing.T) {
	client := NewDefaultClient("http://localhost:8080")

	t.Run("PodsInterface", func(t *testing.T) {
		podInterface := client.Pods()
		assert.NotNil(t, podInterface)
	})

	t.Run("NodesInterface", func(t *testing.T) {
		nodeInterface := client.Nodes()
		assert.NotNil(t, nodeInterface)
	})

	t.Run("ReplicaSetsInterface", func(t *testing.T) {
		rsInterface := client.ReplicaSets()
		assert.NotNil(t, rsInterface)
	})
}

func TestErrorHandling(t *testing.T) {
	// Create test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/pods/not-found":
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("Pod not found"))
			require.NoError(t, err)
		case "/api/v1/pods/server-error":
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("Internal server error"))
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewDefaultClient(server.URL)
	ctx := context.Background()

	t.Run("NotFoundError", func(t *testing.T) {
		_, err := client.Pods().Get(ctx, "not-found")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("ServerError", func(t *testing.T) {
		_, err := client.Pods().Get(ctx, "server-error")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})
}

func TestClientConfiguration(t *testing.T) {
	t.Run("DefaultTimeout", func(t *testing.T) {
		config := Config{
			BaseURL: "http://localhost:8080",
			// Timeout not set, should default to 30s
		}
		client := NewClient(config)
		assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
	})

	t.Run("CustomTimeout", func(t *testing.T) {
		customTimeout := 45 * time.Second
		config := Config{
			BaseURL: "http://localhost:8080",
			Timeout: customTimeout,
		}
		client := NewClient(config)
		assert.Equal(t, customTimeout, client.httpClient.Timeout)
	})
}

func TestBuildURL(t *testing.T) {
	client := NewDefaultClient("http://localhost:8080")

	t.Run("NoParams", func(t *testing.T) {
		url := client.buildURL("/api/v1/pods", nil)
		assert.Equal(t, "http://localhost:8080/api/v1/pods", url)
	})

	t.Run("EmptyParams", func(t *testing.T) {
		url := client.buildURL("/api/v1/pods", map[string]string{})
		assert.Equal(t, "http://localhost:8080/api/v1/pods", url)
	})

	t.Run("WithParams", func(t *testing.T) {
		params := map[string]string{
			"nodeName": "worker-1",
			"status":   "running",
		}
		url := client.buildURL("/api/v1/pods", params)
		assert.Contains(t, url, "http://localhost:8080/api/v1/pods?")
		assert.Contains(t, url, "nodeName=worker-1")
		assert.Contains(t, url, "status=running")
	})
}
