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

func TestReplicaSetOperations(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/replicasets":
			// Create replicaset
			var rs api.ReplicaSet
			err := json.NewDecoder(r.Body).Decode(&rs)
			require.NoError(t, err)

			// Set default status values
			rs.Status = api.ReplicaSetStatus{
				Replicas:      rs.Spec.Replicas,
				ReadyReplicas: 0,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			err = json.NewEncoder(w).Encode(rs)
			require.NoError(t, err)

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/replicasets/test-rs":
			// Get replicaset
			rs := &api.ReplicaSet{
				ObjectMeta: api.ObjectMeta{Name: "test-rs"},
				Spec: api.ReplicaSetSpec{
					Replicas: 3,
					Selector: map[string]string{"app": "test"},
					Template: api.PodTemplateSpec{
						ObjectMeta: api.ObjectMeta{
							Name: "test-pod-template",
						},
						Spec: api.PodSpec{
							Containers: []api.Container{{Name: "test", Image: "nginx"}},
						},
					},
				},
				Status: api.ReplicaSetStatus{
					Replicas:      3,
					ReadyReplicas: 3,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(rs)
			require.NoError(t, err)

		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/replicasets/test-rs":
			// Update replicaset
			var rs api.ReplicaSet
			err := json.NewDecoder(r.Body).Decode(&rs)
			require.NoError(t, err)

			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(rs)
			require.NoError(t, err)

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/replicasets/test-rs":
			// Delete replicaset
			w.WriteHeader(http.StatusNoContent)

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/replicasets":
			// List replicasets
			replicasets := []*api.ReplicaSet{
				{
					ObjectMeta: api.ObjectMeta{Name: "rs1"},
					Spec: api.ReplicaSetSpec{
						Replicas: 2,
						Selector: map[string]string{"app": "app1"},
						Template: api.PodTemplateSpec{
							ObjectMeta: api.ObjectMeta{Name: "template1"},
							Spec: api.PodSpec{
								Containers: []api.Container{{Name: "web", Image: "nginx"}},
							},
						},
					},
					Status: api.ReplicaSetStatus{
						Replicas:      2,
						ReadyReplicas: 2,
					},
				},
				{
					ObjectMeta: api.ObjectMeta{Name: "rs2"},
					Spec: api.ReplicaSetSpec{
						Replicas: 1,
						Selector: map[string]string{"app": "app2"},
						Template: api.PodTemplateSpec{
							ObjectMeta: api.ObjectMeta{Name: "template2"},
							Spec: api.PodSpec{
								Containers: []api.Container{{Name: "api", Image: "alpine"}},
							},
						},
					},
					Status: api.ReplicaSetStatus{
						Replicas:      1,
						ReadyReplicas: 0,
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(replicasets)
			require.NoError(t, err)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewDefaultClient(server.URL)
	ctx := context.Background()

	t.Run("CreateReplicaSet", func(t *testing.T) {
		rs := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "test-rs"},
			Spec: api.ReplicaSetSpec{
				Replicas: 3,
				Selector: map[string]string{"app": "test"},
				Template: api.PodTemplateSpec{
					ObjectMeta: api.ObjectMeta{
						Name: "test-pod-template",
					},
					Spec: api.PodSpec{
						Containers: []api.Container{{Name: "test", Image: "nginx"}},
					},
				},
			},
		}

		result, err := client.ReplicaSets().Create(ctx, rs)
		require.NoError(t, err)
		assert.Equal(t, "test-rs", result.Name)
		assert.Equal(t, int32(3), result.Spec.Replicas)
		assert.Equal(t, int32(3), result.Status.Replicas)
		assert.Equal(t, "test", result.Spec.Selector["app"])
		assert.Equal(t, "nginx", result.Spec.Template.Spec.Containers[0].Image)
	})

	t.Run("GetReplicaSet", func(t *testing.T) {
		rs, err := client.ReplicaSets().Get(ctx, "test-rs")
		require.NoError(t, err)
		assert.Equal(t, "test-rs", rs.Name)
		assert.Equal(t, int32(3), rs.Spec.Replicas)
		assert.Equal(t, int32(3), rs.Status.ReadyReplicas)
		assert.Equal(t, "test", rs.Spec.Selector["app"])
		assert.Equal(t, "test-pod-template", rs.Spec.Template.Name)
	})

	t.Run("UpdateReplicaSet", func(t *testing.T) {
		rs := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "test-rs"},
			Spec: api.ReplicaSetSpec{
				Replicas: 5,
				Selector: map[string]string{"app": "test-updated"},
				Template: api.PodTemplateSpec{
					ObjectMeta: api.ObjectMeta{
						Name: "test-pod-template-updated",
					},
					Spec: api.PodSpec{
						Containers: []api.Container{{Name: "updated", Image: "nginx:latest"}},
					},
				},
			},
			Status: api.ReplicaSetStatus{
				Replicas:      5,
				ReadyReplicas: 3,
			},
		}

		result, err := client.ReplicaSets().Update(ctx, rs)
		require.NoError(t, err)
		assert.Equal(t, "test-rs", result.Name)
		assert.Equal(t, int32(5), result.Spec.Replicas)
		assert.Equal(t, "test-updated", result.Spec.Selector["app"])
		assert.Equal(t, "updated", result.Spec.Template.Spec.Containers[0].Name)
		assert.Equal(t, "nginx:latest", result.Spec.Template.Spec.Containers[0].Image)
	})

	t.Run("DeleteReplicaSet", func(t *testing.T) {
		err := client.ReplicaSets().Delete(ctx, "test-rs")
		require.NoError(t, err)
	})

	t.Run("ListReplicaSets", func(t *testing.T) {
		replicasets, err := client.ReplicaSets().List(ctx)
		require.NoError(t, err)
		assert.Len(t, replicasets, 2)

		// Verify first replicaset
		assert.Equal(t, "rs1", replicasets[0].Name)
		assert.Equal(t, int32(2), replicasets[0].Status.ReadyReplicas)
		assert.Equal(t, "app1", replicasets[0].Spec.Selector["app"])

		// Verify second replicaset
		assert.Equal(t, "rs2", replicasets[1].Name)
		assert.Equal(t, int32(0), replicasets[1].Status.ReadyReplicas)
		assert.Equal(t, "app2", replicasets[1].Spec.Selector["app"])
	})
}

func TestReplicaSetOperationsErrors(t *testing.T) {
	// Create test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/replicasets/not-found":
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("ReplicaSet not found"))
			require.NoError(t, err)
		case "/api/v1/replicasets/server-error":
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("Internal server error"))
			require.NoError(t, err)
		case "/api/v1/replicasets":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusConflict)
				_, err := w.Write([]byte("ReplicaSet already exists"))
				require.NoError(t, err)
			} else if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("Failed to list replicasets"))
				require.NoError(t, err)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewDefaultClient(server.URL)
	ctx := context.Background()

	t.Run("GetReplicaSetNotFound", func(t *testing.T) {
		_, err := client.ReplicaSets().Get(ctx, "not-found")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
		assert.Contains(t, err.Error(), "failed to get replicaset not-found")
	})

	t.Run("GetReplicaSetServerError", func(t *testing.T) {
		_, err := client.ReplicaSets().Get(ctx, "server-error")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("CreateReplicaSetConflict", func(t *testing.T) {
		rs := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "existing-rs"},
		}
		_, err := client.ReplicaSets().Create(ctx, rs)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "409")
		assert.Contains(t, err.Error(), "failed to create replicaset")
	})

	t.Run("ListReplicaSetsError", func(t *testing.T) {
		_, err := client.ReplicaSets().List(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
		assert.Contains(t, err.Error(), "failed to list replicasets")
	})
}

func TestReplicaSetInterface(t *testing.T) {
	client := NewDefaultClient("http://localhost:8080")

	// Verify that ReplicaSets() returns the expected interface
	rsInterface := client.ReplicaSets()
	assert.NotNil(t, rsInterface)
}

func TestReplicaSetScaling(t *testing.T) {
	// Test scaling scenarios with ReplicaSets
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/api/v1/replicasets/scale-test" {
			var rs api.ReplicaSet
			err := json.NewDecoder(r.Body).Decode(&rs)
			require.NoError(t, err)

			// Simulate scaling behavior
			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(rs)
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewDefaultClient(server.URL)
	ctx := context.Background()

	t.Run("ScaleUp", func(t *testing.T) {
		rs := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "scale-test"},
			Spec: api.ReplicaSetSpec{
				Replicas: 10, // Scale up to 10
				Selector: map[string]string{"app": "scale-test"},
			},
			Status: api.ReplicaSetStatus{
				Replicas:      10,
				ReadyReplicas: 8, // 8 ready out of 10
			},
		}

		result, err := client.ReplicaSets().Update(ctx, rs)
		require.NoError(t, err)
		assert.Equal(t, int32(10), result.Spec.Replicas)
		assert.Equal(t, int32(8), result.Status.ReadyReplicas)
	})

	t.Run("ScaleDown", func(t *testing.T) {
		rs := &api.ReplicaSet{
			ObjectMeta: api.ObjectMeta{Name: "scale-test"},
			Spec: api.ReplicaSetSpec{
				Replicas: 2, // Scale down to 2
				Selector: map[string]string{"app": "scale-test"},
			},
			Status: api.ReplicaSetStatus{
				Replicas:      2,
				ReadyReplicas: 2, // All ready
			},
		}

		result, err := client.ReplicaSets().Update(ctx, rs)
		require.NoError(t, err)
		assert.Equal(t, int32(2), result.Spec.Replicas)
		assert.Equal(t, int32(2), result.Status.ReadyReplicas)
	})
}
