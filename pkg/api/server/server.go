package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/emicklei/go-restful/v3"

	"gokube/pkg/api"
	"gokube/pkg/registry"
	"gokube/pkg/storage"
)

// APIServer represents the API server
type APIServer struct {
	nodeRegistry *registry.NodeRegistry
	podRegistry  *registry.PodRegistry
}

// NewAPIServer creates a new instance of APIServer
func NewAPIServer(storage storage.Storage) *APIServer {
	return &APIServer{
		nodeRegistry: registry.NewNodeRegistry(storage),
		podRegistry:  registry.NewPodRegistry(storage),
	}
}

// Start initializes and starts the API server
func (s *APIServer) Start(address string) error {
	container := restful.NewContainer()
	s.registerRoutes(container)

	return http.ListenAndServe(address, container)
}

// registerRoutes adds routes to the container
func (s *APIServer) registerRoutes(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/api/v1").Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/healthz").To(s.healthz))
	// Pod routes
	ws.Route(ws.POST("/pods").To(s.createPod))
	ws.Route(ws.GET("/pods").To(s.listPods))
	ws.Route(ws.GET("/pods/{name}").To(s.getPod))
	ws.Route(ws.PUT("/pods/{name}").To(s.updatePod))
	ws.Route(ws.DELETE("/pods/{name}").To(s.deletePod))
	ws.Route(ws.GET("/pods/unassigned").To(s.listUnassignedPods))

	// Node routes
	ws.Route(ws.POST("/nodes").To(s.createNode))
	ws.Route(ws.GET("/nodes").To(s.listNodes))
	ws.Route(ws.GET("/nodes/{name}").To(s.getNode))
	ws.Route(ws.PUT("/nodes/{name}").To(s.updateNode))
	ws.Route(ws.DELETE("/nodes/{name}").To(s.deleteNode))

	container.Add(ws)
}

func (s *APIServer) healthz(request *restful.Request, response *restful.Response) {
	writeResponse(response, http.StatusOK, nil)
}

// writeResponse is a helper function to write the response and log any errors
func writeResponse(response *restful.Response, status int, entity interface{}) {
	var err error
	if entity != nil {
		err = response.WriteHeaderAndEntity(status, entity)
	} else {
		response.WriteHeader(status)
	}
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// writeError is a helper function to write an error response and log any errors
func writeError(response *restful.Response, status int, err error) {
	writeErr := response.WriteError(status, err)
	if writeErr != nil {
		log.Printf("Error writing error response: %v", writeErr)
	}
}

// createNode handles POST requests to create a new Node
func (s *APIServer) createNode(request *restful.Request, response *restful.Response) {
	node := new(api.Node)
	err := request.ReadEntity(node)
	if err != nil {
		writeError(response, http.StatusBadRequest, err)
		return
	}

	err = s.nodeRegistry.CreateNode(request.Request.Context(), node)
	if err != nil {
		writeError(response, http.StatusInternalServerError, err)
		return
	}

	writeResponse(response, http.StatusCreated, node)
}

// getNode handles GET requests to retrieve a Node
func (s *APIServer) getNode(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	node, err := s.nodeRegistry.GetNode(request.Request.Context(), name)
	if err != nil {
		writeError(response, http.StatusNotFound, err)
		return
	}

	writeResponse(response, http.StatusOK, node)
}

// updateNode handles PUT requests to update a Node
func (s *APIServer) updateNode(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	node := new(api.Node)
	err := request.ReadEntity(node)
	if err != nil {
		writeError(response, http.StatusBadRequest, err)
		return
	}

	if name != node.Name {
		writeError(response, http.StatusBadRequest,
			fmt.Errorf("Node name in URL does not match the name in the request body"))
		return
	}

	err = s.nodeRegistry.UpdateNode(request.Request.Context(), node)
	if err != nil {
		writeError(response, http.StatusInternalServerError, err)
		return
	}

	writeResponse(response, http.StatusOK, node)
}

// deleteNode handles DELETE requests to remove a Node
func (s *APIServer) deleteNode(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	err := s.nodeRegistry.DeleteNode(request.Request.Context(), name)
	if err != nil {
		writeError(response, http.StatusInternalServerError, err)
		return
	}

	writeResponse(response, http.StatusNoContent, nil)
}

// listNodes handles GET requests to list all Nodes
func (s *APIServer) listNodes(request *restful.Request, response *restful.Response) {
	nodes, err := s.nodeRegistry.ListNodes(request.Request.Context())
	if err != nil {
		writeError(response, http.StatusInternalServerError, err)
		return
	}

	writeResponse(response, http.StatusOK, nodes)
}

// createPod handles POST requests to create a new Pod
func (s *APIServer) createPod(request *restful.Request, response *restful.Response) {
	pod := new(api.Pod)
	err := request.ReadEntity(pod)
	if err != nil {
		writeError(response, http.StatusBadRequest, err)
		return
	}

	// Validate Pod spec
	if err := validatePodSpec(pod.Spec); err != nil {
		writeError(response, http.StatusBadRequest, fmt.Errorf("invalid pod spec: %w", err))
		return
	}

	err = s.podRegistry.CreatePod(request.Request.Context(), pod)
	if err != nil {
		writeError(response, http.StatusInternalServerError, err)
		return
	}

	writeResponse(response, http.StatusCreated, pod)
}

// listPods handles GET requests to list all Pods
func (s *APIServer) listPods(request *restful.Request, response *restful.Response) {
	pods, err := s.podRegistry.ListPods(request.Request.Context())
	if err != nil {
		writeError(response, http.StatusInternalServerError, err)
		return
	}

	writeResponse(response, http.StatusOK, pods)
}

// getPod handles GET requests to retrieve a Pod
func (s *APIServer) getPod(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	pod, err := s.podRegistry.GetPod(request.Request.Context(), name)
	if err != nil {
		writeError(response, http.StatusNotFound, err)
		return
	}

	writeResponse(response, http.StatusOK, pod)
}

// updatePod handles PUT requests to update a Pod
func (s *APIServer) updatePod(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	pod := new(api.Pod)
	err := request.ReadEntity(pod)
	if err != nil {
		writeError(response, http.StatusBadRequest, err)
		return
	}

	if name != pod.Name {
		writeError(response, http.StatusBadRequest, fmt.Errorf("pod name in URL does not match pod name in request body"))
		return
	}

	// Validate Pod spec
	if err := validatePodSpec(pod.Spec); err != nil {
		writeError(response, http.StatusBadRequest, fmt.Errorf("invalid pod spec: %w", err))
		return
	}

	err = s.podRegistry.UpdatePod(request.Request.Context(), pod)
	if err != nil {
		writeError(response, http.StatusInternalServerError, err)
		return
	}

	writeResponse(response, http.StatusOK, pod)
}

// deletePod handles DELETE requests to remove a Pod
func (s *APIServer) deletePod(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	err := s.podRegistry.DeletePod(request.Request.Context(), name)
	if err != nil {
		writeError(response, http.StatusInternalServerError, err)
		return
	}

	writeResponse(response, http.StatusNoContent, nil)
}

// listUnassignedPods handles GET requests to list all unassigned Pods
func (s *APIServer) listUnassignedPods(request *restful.Request, response *restful.Response) {
	pods, err := s.podRegistry.ListUnassignedPods(request.Request.Context())
	if err != nil {
		writeError(response, http.StatusInternalServerError, err)
		return
	}

	writeResponse(response, http.StatusOK, pods)
}

func validatePodSpec(spec api.PodSpec) error {
	if len(spec.Containers) == 0 {
		return fmt.Errorf("at least one container must be specified")
	}
	for _, container := range spec.Containers {
		if container.Image == "" {
			return fmt.Errorf("container image must not be empty")
		}
	}
	return nil
}
