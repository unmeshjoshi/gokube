package controller

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"gokube/pkg/api"
	"gokube/pkg/registry"
	"gokube/pkg/registry/names"
)

// ReplicaSetController manages the lifecycle of ReplicaSets
type ReplicaSetController struct {
	replicaSetRegistry *registry.ReplicaSetRegistry
	podRegistry        *registry.PodRegistry
}

// NewReplicaSetController creates a new ReplicaSetController
func NewReplicaSetController(rsRegistry *registry.ReplicaSetRegistry, podRegistry *registry.PodRegistry) *ReplicaSetController {
	return &ReplicaSetController{
		replicaSetRegistry: rsRegistry,
		podRegistry:        podRegistry,
	}
}

func (rsc *ReplicaSetController) Reconcile(ctx context.Context, rs *api.ReplicaSet) error {
	// Get current ReplicaSet state
	currentRS, err := rsc.replicaSetRegistry.Get(ctx, rs.Name)
	if err != nil {
		return err
	}

	// Get all pods
	allPods, err := rsc.podRegistry.ListPods(ctx)
	if err != nil {
		return err
	}

	// Get active pods for this ReplicaSet
	activePods, err := rsc.getPodsForReplicaSet(currentRS, allPods, isPodActiveAndOwnedBy)
	if err != nil {
		return err
	}

	// Compare current pod count with desired replica count
	currentPodCount := len(activePods)
	desiredPodCount := int(currentRS.Spec.Replicas)

	if currentPodCount < desiredPodCount {
		// Create new pods
		for i := currentPodCount; i < desiredPodCount; i++ {
			for _, container := range currentRS.Spec.Template.Spec.Containers {
				pod := &api.Pod{
					ObjectMeta: api.ObjectMeta{
						Name: generatePodNameFromReplicaSet(currentRS.Name),
					},
					Spec: api.PodSpec{
						Containers: []api.Container{container},
					},
				}
				if err := rsc.podRegistry.CreatePod(ctx, pod); err != nil {
					return err
				}
			}
		}
		currentPodCount = desiredPodCount //
		// Update ReplicaSet status
		currentRS.Status.Replicas = int32(currentPodCount)
		return rsc.replicaSetRegistry.Update(ctx, currentRS)

	} else if currentPodCount > desiredPodCount {
		// TODO: Implement pod deletion logic if needed
		currentPodCount = desiredPodCount
		// Update ReplicaSet status
		currentRS.Status.Replicas = int32(currentPodCount)
		return rsc.replicaSetRegistry.Update(ctx, currentRS)
	}

	return nil
}

// GeneratePodNameFromReplicaSet creates a pod name based on the ReplicaSet and container names
func generatePodNameFromReplicaSet(replicaSetName string) string {
	return names.SimpleNameGenerator.GenerateName(replicaSetName)
}

func (rsc *ReplicaSetController) getPodsForReplicaSet(rs *api.ReplicaSet, allPods []*api.Pod, condition func(*api.Pod, *api.ReplicaSet) bool) ([]*api.Pod, error) {
	var activePods []*api.Pod
	for _, pod := range allPods {
		if condition(pod, rs) {
			activePods = append(activePods, pod)
		}
	}

	return activePods, nil
}

func isPodActiveAndOwnedBy(pod *api.Pod, rs *api.ReplicaSet) bool {
	// Check if the pod name contains the ReplicaSet name (ownership)
	return isOwnedBy(pod, rs) && isActive(pod)
}

func (rsc *ReplicaSetController) getPodsOwnedBy(rs *api.ReplicaSet, pods []*api.Pod) ([]*api.Pod, error) {
	return rsc.getPodsForReplicaSet(rs, pods, isOwnedBy)
}

func (rsc *ReplicaSetController) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := rsc.Run(ctx); err != nil {
				fmt.Printf("Error reconciling replicaset: %v\n", err)
			}
		}
	}
}

func (rsc *ReplicaSetController) Run(ctx context.Context) error {

	rscList, err := rsc.replicaSetRegistry.List(context.Background())
	if err != nil {
		log.Fatalf("failed to list replicaSets: %v", err)
		return err
	}

	for _, rs := range rscList {
		err := rsc.Reconcile(context.Background(), rs)
		if err != nil {
			log.Fatalf("failed to reconcile: %v", err)
		}
	}
	return nil
}

func isActive(pod *api.Pod) bool {
	//return pod.Status != api.PodSucceeded && pod.Status != api.PodFailed
	return pod.Status != api.PodFailed //even succeeded pods should be considered active? or else controller keeps on creating pods
}

func isOwnedBy(pod *api.Pod, rs *api.ReplicaSet) bool {
	return strings.HasPrefix(pod.Name, rs.Name)
}
