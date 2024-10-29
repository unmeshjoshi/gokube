package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"gokube/pkg/api"
	"gokube/pkg/api/server"
	"gokube/pkg/controller"
	"gokube/pkg/kubelet"
	"gokube/pkg/registry"
	"gokube/pkg/scheduler"
	"gokube/pkg/storage"
	"google.golang.org/appengine/log"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

func TestGokubeEndToEnd(t *testing.T) {
	cluster := setupTestCluster(t)
	defer cluster.Cleanup()

	rs, err := createReplicaSet(t, cluster)
	if err != nil {
		t.Fatal(err)
	}
	// Wait for the pods to be created
	err = waitForPodCreation(cluster.APIServerURL, rs.Spec.Replicas)
	if err != nil {
		t.Fatalf("Failed to verify pod creation: %v", err)
	}
	t.Log("Verified that 3 pods are created for the ReplicaSet")
	verifyPodsRunning(t, cluster.APIServerURL, rs.Spec.Selector, rs.Spec.Replicas)
}

func createReplicaSet(t *testing.T, cluster *TestCluster) (*api.ReplicaSet, error) {
	// Define a ReplicaSet using the type from your project
	rs := &api.ReplicaSet{
		ObjectMeta: api.ObjectMeta{
			Name: "example-replicaset",
		},
		Spec: api.ReplicaSetSpec{
			Replicas: 3,
			Selector: map[string]string{
				"app": "example-app",
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name: "example-pod",
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}

	// Store the ReplicaSet in the registry
	err := cluster.ReplicaSetRegistry.Create(context.Background(), rs)
	if err != nil {
		t.Fatalf("Failed to create ReplicaSet: %v", err)
	}

	t.Log("ReplicaSet created successfully")
	return rs, err
}

type TestCluster struct {
	EtcdServer         *embed.Etcd
	EtcdClient         *clientv3.Client
	Storage            *storage.EtcdStorage
	ReplicaSetRegistry *registry.ReplicaSetRegistry
	APIServer          *server.APIServer
	APIServerURL       string
	Kubelets           []*kubelet.Kubelet
}

func setupTestCluster(t *testing.T) *TestCluster {
	ctx := context.Background()

	// Start embedded etcd
	etcdServer, _, err := storage.StartEmbeddedEtcd()
	if err != nil {
		t.Fatalf("Failed to start embedded etcd: %v", err)
	}

	// Setup etcd client
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdServer.Config().ListenClientUrls[0].String()},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create etcd client: %v", err)
	}

	// Create storage and registries
	etcdStorage := storage.NewEtcdStorage(etcdClient)
	replicaSetRegistry := registry.NewReplicaSetRegistry(etcdStorage)
	// Create API server
	apiServer := server.NewAPIServer(etcdStorage)

	// Start the API server
	port, err := storage.PickAvailableRandomPort()
	if err != nil {
		t.Fatalf("Failed to pick available random port: %v", err)
	}

	serverURL := "localhost:" + strconv.Itoa(port)
	//TODO: Is this the idiomatic way to handle errors in goroutines?
	go func() {
		err := apiServer.Start(serverURL)
		if err != nil {
			t.Errorf("Failed to start API server: %v", err)
		}
	}()
	// Wait for the API server to be ready
	if err := waitForAPIServer(serverURL); err != nil {
		t.Fatalf("API server failed to start: %v", err)
	}
	t.Log("API Server started at:", serverURL)

	cntr := controller.NewReplicaSetController(replicaSetRegistry, registry.NewPodRegistry(etcdStorage))
	go cntr.Start(ctx)

	schdlr := scheduler.NewScheduler(registry.NewPodRegistry(etcdStorage), registry.NewNodeRegistry(etcdStorage), 1*time.Second)
	go schdlr.Start(ctx)

	kubelets, err := startKubelets(serverURL, 3, t)
	if err != nil {
		t.Fatalf("Failed to start kubelets: %v", err)
	}

	err = waitForKubeletRegistration(serverURL, 3)
	if err != nil {
		t.Fatalf("Kubelet registration failed: %v", err)
	}

	return &TestCluster{
		EtcdServer:         etcdServer,
		EtcdClient:         etcdClient,
		Storage:            etcdStorage,
		APIServer:          apiServer,
		Kubelets:           kubelets,
		ReplicaSetRegistry: replicaSetRegistry,
		APIServerURL:       serverURL,
	}
}

func startKubelets(apiServerIPAndPort string, count int, t *testing.T) ([]*kubelet.Kubelet, error) {
	var kubelets []*kubelet.Kubelet
	for i := 0; i < count; i++ {
		nodeName := fmt.Sprintf("node-%d", i)
		k, err := kubelet.NewKubelet(nodeName, apiServerIPAndPort)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kubelet %s: %v", nodeName, err)
		}
		go func() {
			err := k.Start()
			if err != nil {
				t.Errorf("Failed to start Kubelet %s: %v", nodeName, err)
			}
		}()
		kubelets = append(kubelets, k)
	}
	return kubelets, nil
}

func waitForKubeletRegistration(apiServerURL string, expectedCount int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for Kubelets to register")
		default:
			resp, err := http.Get("http://" + apiServerURL + "/api/v1/nodes")
			if err != nil {
				return fmt.Errorf("failed to list nodes: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			}

			var nodeList []api.Node
			if err := json.NewDecoder(resp.Body).Decode(&nodeList); err != nil {
				return fmt.Errorf("failed to decode node list: %v", err)
			}

			readyCount := 0
			for _, node := range nodeList {
				if node.Status == api.NodeReady {
					readyCount++
				}
			}

			if readyCount == expectedCount {
				return nil
			}

			time.Sleep(1 * time.Second)
		}
	}
}

func (tc *TestCluster) Cleanup() {
	tc.cleanupContainers() //stop etcd after cleanup as cleanup depends on etcd to load metadata about replicasets.
	tc.EtcdClient.Close()
	storage.StopEmbeddedEtcd(tc.EtcdServer)

}

func waitForAPIServer(url string) error {
	for i := 0; i < 30; i++ {
		resp, err := http.Get("http://" + url + "/api/v1/healthz")
		if err == nil && resp.StatusCode == http.StatusOK {

			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("API server did not become ready in time")
}

func waitForPodCreation(apiServerURL string, expectedCount int32) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for pods to be created")
		default:
			resp, err := http.Get("http://" + apiServerURL + "/api/v1/pods")
			if err != nil {
				return fmt.Errorf("failed to list pods: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			}

			var podList []api.Pod
			if err := json.NewDecoder(resp.Body).Decode(&podList); err != nil {
				return fmt.Errorf("failed to decode pod list: %v", err)
			}

			matchingPods := 0
			for _, pod := range podList {
				if matchesSelector(pod) {
					matchingPods++
				}
			}

			if matchingPods == int(expectedCount) {
				return nil
			}

			time.Sleep(1 * time.Second)
		}
	}
}

func matchesSelector(pod api.Pod) bool {
	return strings.Contains(pod.Name, "example-replicaset")
}

func verifyPodsRunning(t *testing.T, apiServerURL string, selector map[string]string, expectedCount int32) {
	err := waitForPodsRunning(apiServerURL, selector, expectedCount)
	if err != nil {
		t.Fatalf("Failed to verify pods running: %v", err)
	}
	t.Logf("Verified that %d pods are running for the ReplicaSet", expectedCount)
}

func waitForPodsRunning(apiServerURL string, selector map[string]string, expectedCount int32) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for pods to be running")
		default:
			pods, err := listPods(apiServerURL)
			if err != nil {
				return fmt.Errorf("failed to list pods: %v", err)
			}

			runningCount := 0
			for _, pod := range pods {
				if matchesSelector(pod) && (pod.Status == api.PodRunning || pod.Status == api.PodSucceeded) {
					runningCount++
				}
			}

			if runningCount == int(expectedCount) {
				return nil
			}

			time.Sleep(5 * time.Second)
		}
	}
}

func (testCluster *TestCluster) cleanupContainers() {
	kubelets := testCluster.Kubelets
	for _, kubelet := range kubelets {
		err := kubelet.CleanupContainers(context.Background())
		if err != nil {
			log.Errorf(context.Background(), "Unable to clean containers for %s", kubelet.GetNodeName())
		}
	}
}

func listPods(apiServerURL string) ([]api.Pod, error) {
	resp, err := http.Get("http://" + apiServerURL + "/api/v1/pods")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var podList []api.Pod
	if err := json.NewDecoder(resp.Body).Decode(&podList); err != nil {
		return nil, fmt.Errorf("failed to decode pod list: %v", err)
	}

	return podList, nil
}
