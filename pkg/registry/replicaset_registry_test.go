package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gokube/pkg/api"
	"gokube/pkg/storage"
)

func createTestReplicaSet(name string, replicas int32, image string) *api.ReplicaSet {
	return &api.ReplicaSet{
		ObjectMeta: api.ObjectMeta{
			Name: name,
		},
		Spec: api.ReplicaSetSpec{
			Replicas: replicas,
			Selector: map[string]string{"app": "test"},
			Template: api.PodTemplateSpec{
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Image: image,
						},
					},
				},
			},
		},
	}
}

func TestReplicaSetRegistry_Create(t *testing.T) {
	ctx := context.Background()
	etcdStorage, etcdServer := setupEtcdStorage()
	defer storage.StopEmbeddedEtcd(etcdServer)

	rs := createTestReplicaSet("test-replicaset", 3, "nginx:latest")
	registry := NewReplicaSetRegistry(etcdStorage)
	err := registry.Create(ctx, rs)
	require.NoError(t, err, "Failed to create ReplicaSet")

	createdRS, err := registry.Get(ctx, "test-replicaset")
	require.NoError(t, err, "Failed to get created ReplicaSet")

	assert.Equal(t, "test-replicaset", createdRS.Name)
	assert.Equal(t, int32(3), createdRS.Spec.Replicas)
	assert.Len(t, createdRS.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, "nginx:latest", createdRS.Spec.Template.Spec.Containers[0].Image)
}

func TestReplicaSetRegistry_Update(t *testing.T) {
	ctx := context.Background()

	etcdStorage, etcdServer := setupEtcdStorage()
	defer storage.StopEmbeddedEtcd(etcdServer)

	rs := createTestReplicaSet("test-replicaset", 3, "nginx:latest")
	registry := NewReplicaSetRegistry(etcdStorage)
	err := registry.Create(ctx, rs)
	require.NoError(t, err, "Failed to create initial ReplicaSet")

	updatedRS := createTestReplicaSet("test-replicaset", 5, "nginx:1.19")
	err = registry.Update(ctx, updatedRS)
	require.NoError(t, err, "Failed to update ReplicaSet")

	retrievedRS, err := registry.Get(ctx, "test-replicaset")
	require.NoError(t, err, "Failed to get updated ReplicaSet")

	assert.Equal(t, int32(5), retrievedRS.Spec.Replicas)
	assert.Len(t, retrievedRS.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, "nginx:1.19", retrievedRS.Spec.Template.Spec.Containers[0].Image)
}

func TestReplicaSetRegistry_List(t *testing.T) {
	ctx := context.Background()

	etcdStorage, etcdServer := setupEtcdStorage()
	defer storage.StopEmbeddedEtcd(etcdServer)

	registry := NewReplicaSetRegistry(etcdStorage)
	rs1 := createTestReplicaSet("test-replicaset-1", 3, "nginx:latest")
	rs2 := createTestReplicaSet("test-replicaset-2", 2, "nginx:1.19")

	require.NoError(t, registry.Create(ctx, rs1))
	require.NoError(t, registry.Create(ctx, rs2))

	rsList, err := registry.List(ctx)
	require.NoError(t, err, "Failed to list ReplicaSets")

	assert.Len(t, rsList, 2)
	assert.Contains(t, []string{rsList[0].Name, rsList[1].Name}, "test-replicaset-1")
	assert.Contains(t, []string{rsList[0].Name, rsList[1].Name}, "test-replicaset-2")
}

func TestReplicaSetRegistry_Delete(t *testing.T) {
	ctx := context.Background()

	etcdStorage, etcdServer := setupEtcdStorage()
	defer storage.StopEmbeddedEtcd(etcdServer)

	registry := NewReplicaSetRegistry(etcdStorage)
	rs := createTestReplicaSet("test-replicaset", 3, "nginx:latest")
	require.NoError(t, registry.Create(ctx, rs))

	err := registry.Delete(ctx, "test-replicaset")
	require.NoError(t, err, "Failed to delete ReplicaSet")

	_, err = registry.Get(ctx, "test-replicaset")
	assert.Error(t, err, "Expected error when getting deleted ReplicaSet")
}

func TestReplicaSetRegistry_GetNonExistent(t *testing.T) {
	ctx := context.Background()

	etcdStorage, etcdServer := setupEtcdStorage()
	defer storage.StopEmbeddedEtcd(etcdServer)

	registry := NewReplicaSetRegistry(etcdStorage)
	_, err := registry.Get(ctx, "non-existent-replicaset")
	assert.Error(t, err, "Expected error when getting non-existent ReplicaSet")
}
