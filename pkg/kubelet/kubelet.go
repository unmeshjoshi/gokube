package kubelet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/emicklei/go-restful/v3"

	"gokube/pkg/api"
	"gokube/pkg/registry/names"
)

type Kubelet struct {
	nodeName     string
	apiServerURL string
	dockerClient *client.Client
	pods         map[string]*api.Pod
}

func NewKubelet(nodeName, apiServerURL string) (*Kubelet, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}

	return &Kubelet{
		nodeName:     nodeName,
		apiServerURL: apiServerURL,
		dockerClient: dockerClient,
		pods:         make(map[string]*api.Pod),
	}, nil
}

func (k *Kubelet) Start() error {
	// Register the node with the API server
	if err := k.registerNode(); err != nil {
		return fmt.Errorf("failed to register node: %w", err)
	}

	// TODO: Implement other Kubelet functionality here

	// Start watching for pod assignments
	go k.watchPods()

	// Start updating pod statuses
	go k.updatePodStatuses()

	return nil
}

func (k *Kubelet) registerNode() error {
	node := &api.Node{
		ObjectMeta: api.ObjectMeta{
			Name: k.nodeName,
		},
		Status: api.NodeReady,
	}

	jsonData, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("failed to marshal node data: %w", err)
	}

	resp, err := http.Post("http://"+k.apiServerURL+"/api/v1/nodes", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request to API server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to register node, status code: %d", resp.StatusCode)
	}

	return nil
}

func (k *Kubelet) watchPods() {
	for {
		pods, err := k.getPodAssignments()
		if err != nil {
			log.Printf("Error getting pod assignments: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if err := k.runNewPods(pods); err != nil {
			log.Printf("Error running new pods: %v", err)
		}

		time.Sleep(10 * time.Second) // Poll every 10 seconds
	}
}

func (k *Kubelet) runNewPods(pods []*api.Pod) error {
	for _, pod := range pods {
		if _, exists := k.pods[pod.Name]; !exists {
			log.Printf("New pod assigned: %s", pod.Name)
			k.pods[pod.Name] = pod
			go k.runPod(pod)
		}
	}
	return nil
}

func (k *Kubelet) getPodAssignments() ([]*api.Pod, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/pods?nodeName=%s", "http://"+k.apiServerURL, k.nodeName))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pods []*api.Pod
	if err := json.NewDecoder(resp.Body).Decode(&pods); err != nil {
		return nil, err
	}

	return pods, nil
}

func (k *Kubelet) runPod(pod *api.Pod) {
	// Simulate running a pod
	log.Printf("Running pod: %s", pod.Name)
	for _, container := range pod.Spec.Containers {
		if err := k.StartContainer(context.Background(), pod, container.Name, container.Image); err != nil {
			log.Printf("Failed to start container %s: %v", container.Name, err)
		}
	}
	// In a real implementation, this would involve setting up containers, etc.
}

func (k *Kubelet) StartContainer(ctx context.Context, pod *api.Pod, containerName, imageName string) error {

	log.Printf("Pulling image: %s", imageName)

	// Pull the image
	out, err := k.dockerClient.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	_, err = io.Copy(os.Stdout, out)
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %v", imageName, err)
	}

	log.Printf("Successfully pulled image: %s", "nginx")

	labels := map[string]string{
		"gokube.pod.name":       pod.Name,
		"gokube.pod.namespace":  pod.Namespace,
		"gokube.container.name": containerName,
	}

	uniqueContainerName := names.SimpleNameGenerator.GenerateName(fmt.Sprintf("%s-%s", pod.Name, containerName))
	// Create the container
	resp, err := k.dockerClient.ContainerCreate(ctx, &container.Config{
		Image:  imageName,
		Labels: labels,
		// You can add more configuration options here as needed
	}, nil, nil, nil, uniqueContainerName)
	if err != nil {
		return fmt.Errorf("failed to create container %s: %v", containerName, err)
	}

	// Start the container
	if err := k.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container %s: %v", containerName, err)
	}

	fmt.Printf("Started container %s with ID %s\n", containerName, resp.ID)
	return nil
}

func (k *Kubelet) GetNodeName() string {
	return k.nodeName
}

type ContainerStatus struct {
	PodName       string
	ContainerName string
	ContainerID   string
	Status        string
}

func (k *Kubelet) ListContainers(ctx context.Context) ([]ContainerStatus, error) {
	containers, err := k.dockerClient.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	var statuses []ContainerStatus
	for _, c := range containers {
		podName, ok := c.Labels["gokube.pod.name"]
		if !ok {
			continue // Skip containers not managed by our system
		}

		pod, ok := k.pods[podName]
		if !ok || pod.NodeName != k.nodeName {
			continue // Skip pods not assigned to this node
		}

		for _, containerSpec := range pod.Spec.Containers {
			if containerSpec.Name == c.Labels["gokube.container.name"] {
				status := ContainerStatus{
					PodName:       podName,
					ContainerName: containerSpec.Name,
					ContainerID:   c.ID,
					Status:        c.State,
				}
				statuses = append(statuses, status)
				break
			}
		}
	}

	return statuses, nil
}

func (k *Kubelet) getPodStatus(ctx context.Context, pod *api.Pod) (api.PodStatus, error) {
	var containerStates []containerState
	for _, container := range pod.Spec.Containers {
		state, err := k.getContainerState(ctx, container.Name)
		if err != nil {
			return api.PodRunning, fmt.Errorf("failed to get state for container %s: %w", container.Name, err)
		}
		containerStates = append(containerStates, state)
	}

	return determinePodStatus(containerStates), nil
}

type containerState struct {
	exists   bool
	running  bool
	exitCode int
}

func (k *Kubelet) getContainerState(ctx context.Context, containerName string) (containerState, error) {
	containerInfo, err := k.dockerClient.ContainerInspect(ctx, containerName)
	if err != nil {
		if client.IsErrNotFound(err) {
			return containerState{exists: false}, nil
		}
		return containerState{}, err
	}

	return containerState{
		exists:   true,
		running:  containerInfo.State.Running,
		exitCode: containerInfo.State.ExitCode,
	}, nil
}

func determinePodStatus(states []containerState) api.PodStatus {
	if anyContainerRunning(states) {
		return api.PodRunning
	}

	if allContainersFailed(states) && anyContainerExists(states) {
		return api.PodFailed
	}

	if allContainersSucceeded(states) {
		return api.PodSucceeded
	}

	return api.PodScheduled
}

func allContainersSucceeded(states []containerState) bool {
	for _, state := range states {
		if state.exitCode != 0 {
			return false
		}
	}
	return true
}

func anyContainerRunning(states []containerState) bool {
	for _, state := range states {
		if state.running {
			return true
		}
	}
	return false
}

func allContainersFailed(states []containerState) bool {
	for _, state := range states {
		if state.exists && state.exitCode == 0 {
			return false
		}
	}
	return true
}

func anyContainerExists(states []containerState) bool {
	for _, state := range states {
		if state.exists {
			return true
		}
	}
	return false
}

func (k *Kubelet) CleanupContainers(ctx context.Context) error {
	containers, err := k.dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("error listing containers for cleanup: %v", err)
	}

	for _, c := range containers {
		if podName, ok := c.Labels["gokube.pod.name"]; ok {
			if pod, exists := k.pods[podName]; exists && pod.NodeName == k.nodeName {
				err := k.dockerClient.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true})
				if err != nil {
					log.Printf("Error removing container %s: %v", c.ID, err)
				} else {
					log.Printf("Removed container %s for pod %s", c.ID, podName)
				}
			}
		}
	}

	return nil
}

func (k *Kubelet) updatePodStatuses() {
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, pod := range k.pods {
				status, err := k.getPodStatus(context.Background(), pod)
				if err != nil {
					log.Printf("Error getting status for pod %s: %v", pod.Name, err)
					continue
				}

				if pod.Status != status {
					pod.Status = status
					if err := k.updatePodStatus(pod); err != nil {
						log.Printf("Error updating status for pod %s: %v", pod.Name, err)
					}
				}
			}
		}
	}
}

func (k *Kubelet) updatePodStatus(pod *api.Pod) error {
	url := fmt.Sprintf("http://%s/api/v1/pods/%s", k.apiServerURL, pod.Name)

	jsonData, err := json.Marshal(pod)
	if err != nil {
		return fmt.Errorf("failed to marshal pod data: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", restful.MIME_JSON)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to API server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update pod status, status code: %d", resp.StatusCode)
	}

	log.Printf("Updated pod status for %s: %v", pod.Name, pod.Status)

	return nil
}
