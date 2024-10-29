package registry

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"go.etcd.io/etcd/server/v3/embed"

	"gokube/pkg/api"
	"gokube/pkg/storage"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func setupEtcdStorage() (storage.Storage, *embed.Etcd) {
	etcdServer, port, err := storage.StartEmbeddedEtcd()
	if err != nil {
		fmt.Printf("Failed to start etcd: %v\n", err)
		os.Exit(1)
	}

	etcdClient, err := clientv3.New(clientv3.Config{
		//TODO: Factor out this in a separate utility which gives client when
		Endpoints:   []string{"http://localhost:" + strconv.Itoa(port)},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		fmt.Printf("Failed to create etcd client: %v\n", err)
		os.Exit(1)
	}

	etcdStorage := storage.NewEtcdStorage(etcdClient)
	return etcdStorage, etcdServer
}

func TestCreateAndUpdatePod(t *testing.T) {
	etcdStorage, etcdServer := setupEtcdStorage()

	defer storage.StopEmbeddedEtcd(etcdServer)

	registry := NewPodRegistry(etcdStorage)

	ctx := context.Background()

	// Test Create
	pod := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name: "test-pod",
		},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{
					Image: "nginx:latest",
				},
			},
			Replicas: 3,
		},
		Status: api.PodPending,
	}

	err := registry.CreatePod(ctx, pod)
	if err != nil {
		t.Fatalf("Failed to create pod: %v", err)
	}

	// Verify pod was created
	createdPod, err := registry.GetPod(ctx, "test-pod")
	if err != nil {
		t.Fatalf("Failed to get created pod: %v", err)
	}
	if createdPod.Name != "test-pod" || createdPod.Status != api.PodPending {
		t.Errorf("Created pod does not match expected: got %v, want %v", createdPod, pod)
	}
	if createdPod.Spec.Replicas != 3 || len(createdPod.Spec.Containers) != 1 || createdPod.Spec.Containers[0].Image != "nginx:latest" {
		t.Errorf("Created pod spec does not match expected: got %v, want %v", createdPod.Spec, pod.Spec)
	}

	// Test Update
	updatedPod := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name: "test-pod",
		},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{
					Image: "nginx:1.19",
				},
			},
			Replicas: 5,
		},
		Status: api.PodPending,
	}

	err = registry.UpdatePod(ctx, updatedPod)
	if err != nil {
		t.Fatalf("Failed to update pod: %v", err)
	}

	// Verify pod was updated
	retrievedPod, err := registry.GetPod(ctx, "test-pod")
	if err != nil {
		t.Fatalf("Failed to get updated pod: %v", err)
	}
	if retrievedPod.Status != api.PodPending {
		t.Errorf("Updated pod status does not match expected: got %v, want %v", retrievedPod.Status, api.PodPending)
	}
	if retrievedPod.Spec.Replicas != 5 || len(retrievedPod.Spec.Containers) != 1 || retrievedPod.Spec.Containers[0].Image != "nginx:1.19" {
		t.Errorf("Updated pod spec does not match expected: got %v, want %v", retrievedPod.Spec, updatedPod.Spec)
	}

	// Clean up
	err = registry.DeletePod(ctx, "test-pod")
	if err != nil {
		t.Fatalf("Failed to delete pod: %v", err)
	}

	// Test cases
	testCases := []struct {
		name            string
		podsToCreate    []*api.Pod
		expectedPending int
	}{
		{
			name: "No pending pods",
			podsToCreate: []*api.Pod{
				{ObjectMeta: api.ObjectMeta{Name: "pod1"},
					Spec:   api.PodSpec{Containers: []api.Container{{Name: "test-container2", Image: "nginx"}}},
					Status: api.PodRunning},
				{ObjectMeta: api.ObjectMeta{Name: "pod2"},
					Spec:   api.PodSpec{Containers: []api.Container{{Name: "test-container2", Image: "nginx"}}},
					Status: api.PodRunning},
			},
			expectedPending: 0,
		},
		{
			name: "Some pending pods",
			podsToCreate: []*api.Pod{
				{ObjectMeta: api.ObjectMeta{Name: "pod3"},
					Spec:   api.PodSpec{Containers: []api.Container{{Name: "test-container2", Image: "nginx"}}},
					Status: api.PodPending},
				{ObjectMeta: api.ObjectMeta{Name: "pod4"},
					Spec:   api.PodSpec{Containers: []api.Container{{Name: "test-container2", Image: "nginx"}}},
					Status: api.PodRunning},
				{ObjectMeta: api.ObjectMeta{Name: "pod5"},
					Spec:   api.PodSpec{Containers: []api.Container{{Name: "test-container2", Image: "nginx"}}},
					Status: api.PodPending},
			},
			expectedPending: 2,
		},
		{
			name: "All pending pods",
			podsToCreate: []*api.Pod{
				{ObjectMeta: api.ObjectMeta{Name: "pod6"},
					Spec:   api.PodSpec{Containers: []api.Container{{Name: "test-container2", Image: "nginx"}}},
					Status: api.PodPending},
				{ObjectMeta: api.ObjectMeta{Name: "pod7"},
					Spec:   api.PodSpec{Containers: []api.Container{{Name: "test-container2", Image: "nginx"}}},
					Status: api.PodPending},
			},
			expectedPending: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			// Clean up before each test
			if err := etcdStorage.DeletePrefix(ctx, podPrefix); err != nil {
				t.Fatalf("Failed to clean up Pods: %v", err)
			}

			// Create test pods
			for _, pod := range tc.podsToCreate {
				if err := registry.CreatePod(ctx, pod); err != nil {
					t.Fatalf("Failed to create test pod: %v", err)
				}
			}

			// Call ListPendingPods
			pendingPods, err := registry.ListPendingPods(ctx)
			if err != nil {
				t.Fatalf("ListPendingPods failed: %v", err)
			}

			// Check the number of pending pods
			if len(pendingPods) != tc.expectedPending {
				t.Errorf("Expected %d pending pods, but got %d", tc.expectedPending, len(pendingPods))
			}

			// Check that all returned pods are actually pending
			for _, pod := range pendingPods {
				if pod.Status != api.PodPending {
					t.Errorf("Pod %s is not pending, status: %s", pod.Name, pod.Status)
				}
			}
		})
	}
}
