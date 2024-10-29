package scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"gokube/pkg/api"
	"gokube/pkg/registry"
)

type Scheduler struct {
	podRegistry    *registry.PodRegistry
	nodeRegistry   *registry.NodeRegistry
	schedulingRate time.Duration
}

func NewScheduler(podRegistry *registry.PodRegistry, nodeRegistry *registry.NodeRegistry, schedulingRate time.Duration) *Scheduler {
	return &Scheduler{
		podRegistry:    podRegistry,
		nodeRegistry:   nodeRegistry,
		schedulingRate: schedulingRate,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.schedulingRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.schedulePendingPods(ctx); err != nil {
				fmt.Printf("Error scheduling pods: %v\n", err)
			}
		}
	}
}

func (s *Scheduler) schedulePendingPods(ctx context.Context) error {
	// Get all pending pods
	pods, err := s.podRegistry.ListPendingPods(ctx)
	if err != nil {
		return fmt.Errorf("failed to list pending pods: %v", err)
	}

	// Get all available nodes
	nodes, err := s.nodeRegistry.ListNodes(ctx)
	if err != nil {
		return fmt.Errorf("failed to list nodes: %v", err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes available for scheduling")
	}

	// Simple round-robin scheduling
	for _, pod := range pods {
		//TODO: We are picking up the nodes randomly. Need to have better policy based on the node status that the kubelet
		//		updates periodically
		node := nodes[rand.Intn(len(nodes))]

		// Assign the pod to the node
		pod.NodeName = node.Name
		pod.Status = api.PodScheduled

		// Update the pod in the registry
		if err := s.podRegistry.UpdatePod(ctx, pod); err != nil {
			return fmt.Errorf("failed to update pod %s: %v", pod.Name, err)
		}

		fmt.Printf("Scheduled pod %s on node %s\n", pod.Name, node.Name)
	}

	return nil
}
