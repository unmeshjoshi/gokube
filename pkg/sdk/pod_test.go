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

func TestPodOperations(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/pods":
			// Create pod
			var pod api.Pod
			err := json.NewDecoder(r.Body).Decode(&pod)
			require.NoError(t, err)

			pod.Status = api.PodPending
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(pod)

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/pods/test-pod":
			// Get pod
			pod := &api.Pod{
				ObjectMeta: api.ObjectMeta{Name: "test-pod"},
				Spec: api.PodSpec{
					Containers: []api.Container{{Name: "test", Image: "nginx"}},
				},
				Status: api.PodRunning,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(pod)

		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/pods/test-pod":
			// Update pod
			var pod api.Pod
			err := json.NewDecoder(r.Body).Decode(&pod)
			require.NoError(t, err)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(pod)

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/pods/test-pod":
			// Delete pod
			w.WriteHeader(http.StatusNoContent)

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/pods":
			// List pods
			pods := []*api.Pod{
				{
					ObjectMeta: api.ObjectMeta{Name: "pod1"},
					Status:     api.PodRunning,
					NodeName:   "node1",
				},
				{
					ObjectMeta: api.ObjectMeta{Name: "pod2"},
					Status:     api.PodPending,
				},
			}

			// Filter by nodeName if provided
			nodeName := r.URL.Query().Get("nodeName")
			if nodeName != "" {
				filteredPods := []*api.Pod{}
				for _, pod := range pods {
					if pod.NodeName == nodeName {
						filteredPods = append(filteredPods, pod)
					}
				}
				pods = filteredPods
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(pods)

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/pods/unassigned":
			// List unassigned pods
			pods := []*api.Pod{
				{
					ObjectMeta: api.ObjectMeta{Name: "unassigned-pod"},
					Status:     api.PodPending,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(pods)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewDefaultClient(server.URL)
	ctx := context.Background()

	t.Run("CreatePod", func(t *testing.T) {
		pod := &api.Pod{
			ObjectMeta: api.ObjectMeta{Name: "test-pod"},
			Spec: api.PodSpec{
				Containers: []api.Container{{Name: "test", Image: "nginx"}},
			},
		}

		result, err := client.Pods().Create(ctx, pod)
		require.NoError(t, err)
		assert.Equal(t, "test-pod", result.Name)
		assert.Equal(t, api.PodPending, result.Status)
	})

	t.Run("GetPod", func(t *testing.T) {
		pod, err := client.Pods().Get(ctx, "test-pod")
		require.NoError(t, err)
		assert.Equal(t, "test-pod", pod.Name)
		assert.Equal(t, api.PodRunning, pod.Status)
		assert.Equal(t, "test", pod.Spec.Containers[0].Name)
		assert.Equal(t, "nginx", pod.Spec.Containers[0].Image)
	})

	t.Run("UpdatePod", func(t *testing.T) {
		pod := &api.Pod{
			ObjectMeta: api.ObjectMeta{Name: "test-pod"},
			Spec: api.PodSpec{
				Containers: []api.Container{{Name: "updated", Image: "nginx:latest"}},
			},
			Status:   api.PodRunning,
			NodeName: "worker-node-1",
		}

		result, err := client.Pods().Update(ctx, pod)
		require.NoError(t, err)
		assert.Equal(t, "test-pod", result.Name)
		assert.Equal(t, "updated", result.Spec.Containers[0].Name)
		assert.Equal(t, "nginx:latest", result.Spec.Containers[0].Image)
		assert.Equal(t, api.PodRunning, result.Status)
		assert.Equal(t, "worker-node-1", result.NodeName)
	})

	t.Run("DeletePod", func(t *testing.T) {
		err := client.Pods().Delete(ctx, "test-pod")
		require.NoError(t, err)
	})

	t.Run("ListPods", func(t *testing.T) {
		pods, err := client.Pods().List(ctx)
		require.NoError(t, err)
		assert.Len(t, pods, 2)
		assert.Equal(t, "pod1", pods[0].Name)
		assert.Equal(t, "pod2", pods[1].Name)
		assert.Equal(t, api.PodRunning, pods[0].Status)
		assert.Equal(t, api.PodPending, pods[1].Status)
	})

	t.Run("ListPodsWithNodeFilter", func(t *testing.T) {
		pods, err := client.Pods().List(ctx, WithNodeName("node1"))
		require.NoError(t, err)
		assert.Len(t, pods, 1)
		assert.Equal(t, "pod1", pods[0].Name)
		assert.Equal(t, "node1", pods[0].NodeName)
	})

	t.Run("ListPodsWithNonExistentNode", func(t *testing.T) {
		pods, err := client.Pods().List(ctx, WithNodeName("non-existent"))
		require.NoError(t, err)
		assert.Len(t, pods, 0)
	})

	t.Run("ListUnassignedPods", func(t *testing.T) {
		pods, err := client.Pods().ListUnassigned(ctx)
		require.NoError(t, err)
		assert.Len(t, pods, 1)
		assert.Equal(t, "unassigned-pod", pods[0].Name)
		assert.Equal(t, api.PodPending, pods[0].Status)
	})
}

func TestPodOperationsErrors(t *testing.T) {
	// Create test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/pods/not-found":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Pod not found"))
		case "/api/v1/pods/server-error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
		case "/api/v1/pods":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte("Pod already exists"))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewDefaultClient(server.URL)
	ctx := context.Background()

	t.Run("GetPodNotFound", func(t *testing.T) {
		_, err := client.Pods().Get(ctx, "not-found")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Contains(t, err.Error(), "failed to get pod not-found")
	})

	t.Run("GetPodServerError", func(t *testing.T) {
		_, err := client.Pods().Get(ctx, "server-error")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("CreatePodConflict", func(t *testing.T) {
		pod := &api.Pod{
			ObjectMeta: api.ObjectMeta{Name: "existing-pod"},
		}
		_, err := client.Pods().Create(ctx, pod)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "409")
		assert.Contains(t, err.Error(), "failed to create pod")
	})
}

func TestPodInterface(t *testing.T) {
	client := NewDefaultClient("http://localhost:8080")

	// Verify that Pods() returns the expected interface
	podInterface := client.Pods()
	assert.NotNil(t, podInterface)

	// Type assertion to ensure it implements PodInterface
	_, ok := podInterface.(PodInterface)
	assert.True(t, ok, "Pods() should return an implementation of PodInterface")
}
