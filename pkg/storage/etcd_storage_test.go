package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"

	"go.etcd.io/etcd/api/v3/mvccpb"
)

func TestEmbeddedEtcd(t *testing.T) {
	t.Run("should be able to put and get key", func(t *testing.T) {
		TestWithEmbeddedEtcd(t, func(t *testing.T, cli *clientv3.Client) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := cli.Put(ctx, "test-key", "test-value")
			assert.NoError(t, err)

			resp, err := cli.Get(ctx, "test-key")
			assert.NoError(t, err)

			assert.Len(t, resp.Kvs, 1)
			assert.Equal(t, "test-value", string(resp.Kvs[0].Value))
		})
	})

	t.Run("should be able to update existing key", func(t *testing.T) {
		TestWithEmbeddedEtcd(t, func(t *testing.T, cli *clientv3.Client) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := cli.Put(ctx, "test-key", "test-value")
			assert.NoError(t, err)

			_, err = cli.Put(ctx, "test-key", "updated-value")
			assert.NoError(t, err)

			resp, err := cli.Get(ctx, "test-key")
			assert.NoError(t, err)

			assert.Equal(t, "updated-value", string(resp.Kvs[0].Value))
		})
	})

	t.Run("should be able to delete key", func(t *testing.T) {
		TestWithEmbeddedEtcd(t, func(t *testing.T, cli *clientv3.Client) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := cli.Put(ctx, "test-key", "test-value")
			assert.NoError(t, err)

			_, err = cli.Delete(ctx, "test-key")
			assert.NoError(t, err)

			resp, err := cli.Get(ctx, "test-key")
			assert.NoError(t, err)

			assert.Len(t, resp.Kvs, 0)
		})
	})

	t.Run("should be able to list keys with prefix", func(t *testing.T) {
		TestWithEmbeddedEtcd(t, func(t *testing.T, cli *clientv3.Client) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := cli.Put(ctx, "/prefix/key1", "value1")
			assert.NoError(t, err)

			_, err = cli.Put(ctx, "/prefix/key2", "value2")
			assert.NoError(t, err)

			resp, err := cli.Get(ctx, "/prefix/", clientv3.WithPrefix())
			assert.NoError(t, err)

			assert.Len(t, resp.Kvs, 2)
		})
	})

	t.Run("should be able to watch key", func(t *testing.T) {
		TestWithEmbeddedEtcd(t, func(t *testing.T, cli *clientv3.Client) {
			watchKey := "/watch-test/key"
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			watchChan := cli.Watch(ctx, watchKey)

			go func() {
				time.Sleep(1 * time.Second)
				_, err := cli.Put(ctx, watchKey, "initial-value")
				assert.NoError(t, err)

				time.Sleep(1 * time.Second)
				_, err = cli.Put(ctx, watchKey, "updated-value")
				assert.NoError(t, err)

				time.Sleep(1 * time.Second)
				_, err = cli.Delete(ctx, watchKey)
				assert.NoError(t, err)
			}()

			expectedEvents := []struct {
				Type  mvccpb.Event_EventType
				Value string
			}{
				{mvccpb.PUT, "initial-value"},
				{mvccpb.PUT, "updated-value"},
				{mvccpb.DELETE, ""},
			}

			for _, expected := range expectedEvents {
				select {
				case watchResp := <-watchChan:
					assert.Len(t, watchResp.Events, 1)

					ev := watchResp.Events[0]
					assert.Equal(t, expected.Type, ev.Type)
					assert.Equal(t, expected.Value, string(ev.Kv.Value))
				case <-ctx.Done():
					t.Fatalf("Watch timed out")
				}
			}
		})
	})
}
