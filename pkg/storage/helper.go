package storage

import (
	"fmt"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// TestWithEmbeddedEtcd takes in testing.T, starts the embedded etcd server
// handles the cleanup of the server after the test is done and invokes the test function
// with the embedded etcd server instance
func TestWithEmbeddedEtcd(t *testing.T, test func(t *testing.T, etcdServer *clientv3.Client)) {
	etcdServer, port, err := StartEmbeddedEtcd()
	if err != nil {
		t.Fatalf("Failed to start embedded etcd: %v", err)
	}
	defer StopEmbeddedEtcd(etcdServer)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{fmt.Sprintf("http://localhost:%d", port)},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		t.Fatalf("Failed to create etcd client: %v", err)
	}
	defer func() {
		if err := cli.Close(); err != nil {
			t.Fatalf("Failed to close etcd client: %v", err)
		}
	}()

	test(t, cli)
}
