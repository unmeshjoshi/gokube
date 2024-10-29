package controller

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"

	"gokube/pkg/api"
	"gokube/pkg/registry"
	"gokube/pkg/storage"
)

var rsRegistry *registry.ReplicaSetRegistry
var podRegistry *registry.PodRegistry
var rsController *ReplicaSetController
var etcdClient *clientv3.Client
var etcdServer *embed.Etcd

func setup() {
	var err error
	var port int
	etcdServer, port, err = storage.StartEmbeddedEtcd()
	if err != nil {
		fmt.Printf("Failed to start etcd: %v\n", err)
		os.Exit(1)
	}
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://localhost:" + strconv.Itoa(port)},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Printf("Failed to create etcd client: %v\n", err)
		os.Exit(1)
	}

	etcdStorage := storage.NewEtcdStorage(etcdClient)
	podRegistry = registry.NewPodRegistry(etcdStorage)
	rsRegistry = registry.NewReplicaSetRegistry(etcdStorage)
	rsController = NewReplicaSetController(rsRegistry, podRegistry)
}

func teardown() {
	storage.StopEmbeddedEtcd(etcdServer)
}

func TestReconcile(t *testing.T) {
	setup()
	defer teardown()

	// Create storage and registries
	etcdStorage := storage.NewEtcdStorage(etcdClient)
	replicaSetRegistry := registry.NewReplicaSetRegistry(etcdStorage)
	podRegistry := registry.NewPodRegistry(etcdStorage)

	// Create ReplicaSetController
	rsc := NewReplicaSetController(replicaSetRegistry, podRegistry)

	testCases := []struct {
		name          string
		initialRS     *api.ReplicaSet
		initialPods   []*api.Pod
		expectedPods  int
		expectedError bool
	}{
		{
			name: "Create pods when fewer than desired",
			initialRS: &api.ReplicaSet{
				ObjectMeta: api.ObjectMeta{Name: "test-rs-1"},
				Spec: api.ReplicaSetSpec{
					Replicas: 3,
					Template: api.PodTemplateSpec{
						Spec: api.PodSpec{
							Containers: []api.Container{{Name: "test-container", Image: "nginx"}},
						},
					},
				},
			},
			initialPods:   []*api.Pod{},
			expectedPods:  3,
			expectedError: false,
		},
		{
			name: "Do nothing when pod count matches desired",
			initialRS: &api.ReplicaSet{
				ObjectMeta: api.ObjectMeta{Name: "test-rs-2"},
				Spec: api.ReplicaSetSpec{
					Replicas: 2,
					Template: api.PodTemplateSpec{
						Spec: api.PodSpec{
							Containers: []api.Container{{Name: "test-container", Image: "nginx"}},
						},
					},
				},
				Status: api.ReplicaSetStatus{Replicas: 2},
			},
			initialPods: []*api.Pod{
				{ObjectMeta: api.ObjectMeta{Name: "test-rs-2-test-container-1"}, Spec: api.PodSpec{
					Containers: []api.Container{{Name: "test-container1", Image: "nginx"}},
				}},
				{ObjectMeta: api.ObjectMeta{Name: "test-rs-2-test-container-2"}, Spec: api.PodSpec{
					Containers: []api.Container{{Name: "test-container2", Image: "nginx"}},
				}},
			},
			expectedPods:  2,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			err := replicaSetRegistry.Delete(ctx, tc.initialRS.Name)
			if err != nil {
				t.Fatalf("Failed to Delete ReplicaSet: %v", err)
			}
			// Create initial ReplicaSet
			if err := replicaSetRegistry.Create(ctx, tc.initialRS); err != nil {
				t.Fatalf("Failed to create initial ReplicaSet: %v", err)
			}

			// Create initial Pods
			for _, pod := range tc.initialPods {
				if err := podRegistry.CreatePod(ctx, pod); err != nil {
					t.Fatalf("Failed to create initial Pod: %v", err)
				}
			}

			// Run Reconcile
			err = rsc.Reconcile(ctx, tc.initialRS)

			if tc.expectedError && err == nil {
				t.Error("Expected an error, but got none")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check the number of pods
			allPods, err := podRegistry.ListPods(ctx)
			if err != nil {
				t.Fatalf("Failed to list pods: %v", err)
			}
			actualPods, err := rsc.getPodsOwnedBy(tc.initialRS, allPods)
			if err != nil {
				t.Fatalf("Failed to list pods: %v", err)
			}
			if len(actualPods) != tc.expectedPods {
				t.Errorf("Expected %d pods, but got %d", tc.expectedPods, len(actualPods))
			}

			// Check the ReplicaSet status
			updatedRS, err := replicaSetRegistry.Get(ctx, tc.initialRS.Name)
			if err != nil {
				t.Fatalf("Failed to get updated ReplicaSet: %v", err)
			}
			if updatedRS.Status.Replicas != int32(len(actualPods)) {
				t.Errorf("Expected ReplicaSet status to be updated to %d, but got %d", len(actualPods), updatedRS.Status.Replicas)
			}
		})
	}
}

func TestGetActivePodsForReplicaSet(t *testing.T) {
	rs := &api.ReplicaSet{
		ObjectMeta: api.ObjectMeta{
			Name: "test-rs",
		},
	}

	testCases := []struct {
		name          string
		pods          []*api.Pod
		expectedCount int
	}{
		{
			name: "All active and owned pods",
			pods: []*api.Pod{
				{ObjectMeta: api.ObjectMeta{Name: "test-rs-pod1"}, Status: api.PodRunning},
				{ObjectMeta: api.ObjectMeta{Name: "test-rs-pod2"}, Status: api.PodPending},
			},
			expectedCount: 2,
		},
		{
			name: "Mix of active, inactive, and unowned pods",
			pods: []*api.Pod{
				{ObjectMeta: api.ObjectMeta{Name: "test-rs-pod1"}, Status: api.PodRunning},
				{ObjectMeta: api.ObjectMeta{Name: "test-rs-pod2"}, Status: api.PodSucceeded},
				{ObjectMeta: api.ObjectMeta{Name: "test-rs-pod3"}, Status: api.PodFailed},
				{ObjectMeta: api.ObjectMeta{Name: "other-rs-pod"}, Status: api.PodRunning},
			},
			expectedCount: 2, //succeeded is considered active FIXME:
		},
		{
			name:          "No pods",
			pods:          []*api.Pod{},
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var rsc = &ReplicaSetController{}
			activePods, err := rsc.getPodsForReplicaSet(rs, tc.pods, isPodActiveAndOwnedBy)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(activePods) != tc.expectedCount {
				t.Errorf("Expected %d active pods, got %d", tc.expectedCount, len(activePods))
			}

			for _, pod := range activePods {
				if (pod.Status != api.PodRunning && pod.Status != api.PodSucceeded) && pod.Status != api.PodPending {
					t.Errorf("Expected pod status to be Running/Succeeded or Pending, got %s", pod.Status)
				}
				if len(pod.Name) <= len(rs.Name) || pod.Name[:len(rs.Name)] != rs.Name {
					t.Errorf("Expected pod name to start with %s, got %s", rs.Name, pod.Name)
				}
			}
		})
	}
}

// Other necessary stub methods...
