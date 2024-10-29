package scheduler

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gokube/pkg/api"
	"gokube/pkg/registry"
	"gokube/pkg/storage"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestScheduler_SchedulePendingPods(t *testing.T) {
	// Start embedded etcd
	etcdServer, port, err := storage.StartEmbeddedEtcd()
	if err != nil {
		t.Fatalf("Failed to start embedded etcd: %v", err)
	}
	defer etcdServer.Close()

	// Setup etcd client
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{fmt.Sprintf("localhost:%d", port)},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create etcd client: %v", err)
	}
	defer etcdClient.Close()

	// Create storage and registries
	etcdStorage := storage.NewEtcdStorage(etcdClient)
	podRegistry := registry.NewPodRegistry(etcdStorage)
	nodeRegistry := registry.NewNodeRegistry(etcdStorage)

	// Create scheduler
	scheduler := NewScheduler(podRegistry, nodeRegistry, 1*time.Second)

	// Test cases
	testCases := []struct {
		name              string
		nodes             []*api.Node
		pendingPods       []*api.Pod
		expectedScheduled int
	}{
		{
			name: "Schedule pending pods to available nodes",
			nodes: []*api.Node{
				{ObjectMeta: api.ObjectMeta{Name: "node1"}},
				{ObjectMeta: api.ObjectMeta{Name: "node2"}},
			},
			pendingPods: []*api.Pod{
				{
					ObjectMeta: api.ObjectMeta{Name: "pod1"},
					Spec: api.PodSpec{
						Containers: []api.Container{{Name: "container1", Image: "nginx:latest"}},
					},
					Status: api.PodPending,
				},
				{
					ObjectMeta: api.ObjectMeta{Name: "pod2"},
					Spec: api.PodSpec{
						Containers: []api.Container{{Name: "container2", Image: "redis:latest"}},
					},
					Status: api.PodPending,
				},
				{
					ObjectMeta: api.ObjectMeta{Name: "pod3"},
					Spec: api.PodSpec{
						Containers: []api.Container{{Name: "container3", Image: "mysql:5.7"}},
					},
					Status: api.PodPending,
				},
			},
			expectedScheduled: 3,
		},
		{
			name:  "No nodes available",
			nodes: []*api.Node{},
			pendingPods: []*api.Pod{
				{
					ObjectMeta: api.ObjectMeta{Name: "pod4"},
					Spec: api.PodSpec{
						Containers: []api.Container{{Name: "container4", Image: "busybox:latest"}},
					},
					Status: api.PodPending,
				},
			},
			expectedScheduled: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			// Clean up before each test
			if err := etcdStorage.DeletePrefix(ctx, "/pods/"); err != nil {
				t.Fatalf("Failed to clean up Pods: %v", err)
			}
			if err := etcdStorage.DeletePrefix(ctx, "/registry/nodes/"); err != nil {
				t.Fatalf("Failed to clean up Nodes: %v", err)
			}

			// Create test nodes
			for _, node := range tc.nodes {
				if err := nodeRegistry.CreateNode(ctx, node); err != nil {
					t.Fatalf("Failed to create test node: %v", err)
				}
			}

			// Create pending pods
			for _, pod := range tc.pendingPods {
				if err := podRegistry.CreatePod(ctx, pod); err != nil {
					t.Fatalf("Failed to create test pod: %v", err)
				}
			}

			// Run scheduler
			if err := scheduler.schedulePendingPods(ctx); err != nil {
				if tc.expectedScheduled > 0 {
					t.Fatalf("Scheduler failed: %v", err)
				}
			}

			// Check scheduled pods
			scheduledPods, err := podRegistry.ListPods(ctx)
			if err != nil {
				t.Fatalf("Failed to list pods: %v", err)
			}

			scheduledCount := 0
			for _, pod := range scheduledPods {
				if pod.Status == api.PodScheduled && pod.NodeName != "" {
					scheduledCount++
				}
			}

			if scheduledCount != tc.expectedScheduled {
				t.Errorf("Expected %d scheduled pods, but got %d", tc.expectedScheduled, scheduledCount)
			}
		})
	}
}
