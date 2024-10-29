package registry

import (
	"context"
	"fmt"
	"sync"

	"gokube/pkg/api"
	"gokube/pkg/storage"
)

type ReplicaSetRegistry struct {
	storage storage.Storage
	mutex   sync.RWMutex
}

func NewReplicaSetRegistry(storage storage.Storage) *ReplicaSetRegistry {
	return &ReplicaSetRegistry{
		storage: storage,
	}
}

func (r *ReplicaSetRegistry) Create(ctx context.Context, rs *api.ReplicaSet) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	key := fmt.Sprintf("/replicasets/%s", rs.Name)

	// Check if ReplicaSet already exists
	existingRS := &api.ReplicaSet{}
	err := r.storage.Get(ctx, key, existingRS)
	if err == nil {
		return fmt.Errorf("replicaset %s already exists", rs.Name)
	}

	// Store the ReplicaSet
	return r.storage.Create(ctx, key, rs)
}

func (r *ReplicaSetRegistry) Get(ctx context.Context, name string) (*api.ReplicaSet, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	key := fmt.Sprintf("/replicasets/%s", name)
	rs := &api.ReplicaSet{}
	err := r.storage.Get(ctx, key, rs)
	if err != nil {
		return nil, fmt.Errorf("replicaset %s not found: %v", name, err)
	}

	return rs, nil
}

func (r *ReplicaSetRegistry) Update(ctx context.Context, rs *api.ReplicaSet) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	key := fmt.Sprintf("/replicasets/%s", rs.Name)

	// Check if ReplicaSet exists
	existingRS := &api.ReplicaSet{}
	err := r.storage.Get(ctx, key, existingRS)
	if err != nil {
		return fmt.Errorf("replicaset %s not found: %v", rs.Name, err)
	}

	// Update the ReplicaSet
	return r.storage.Update(ctx, key, rs)
}

func (r *ReplicaSetRegistry) Delete(ctx context.Context, name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	key := fmt.Sprintf("/replicasets/%s", name)
	return r.storage.Delete(ctx, key)
}

func (r *ReplicaSetRegistry) List(ctx context.Context) ([]*api.ReplicaSet, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	prefix := "/replicasets/"
	var replicaSets []*api.ReplicaSet

	err := r.storage.List(ctx, prefix, &replicaSets)
	if err != nil {
		return nil, fmt.Errorf("error listing replicasets: %v", err)
	}

	return replicaSets, nil
}
