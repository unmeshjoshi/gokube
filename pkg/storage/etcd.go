package storage

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/etcd/server/v3/embed"
)

func StartEmbeddedEtcd() (*embed.Etcd, int, error) {
	cfg := embed.NewConfig()
	dir, err := createTempDir()
	if err != nil {
		return nil, 0, err
	}

	cfg.Dir = dir

	// Use a random available port
	port, err := PickAvailableRandomPort()
	if err != nil {
		return nil, 0, err
	}

	cfg.ListenPeerUrls = []url.URL{{Scheme: "http", Host: fmt.Sprintf("127.0.0.1:%d", port)}}

	port, err = PickAvailableRandomPort()
	if err != nil {
		return nil, port, err
	}
	cfg.ListenClientUrls = []url.URL{{Scheme: "http", Host: fmt.Sprintf("127.0.0.1:%d", port)}}

	cfg.Logger = "zap"
	cfg.LogOutputs = []string{"stderr"}

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		return nil, 0, err
	}

	select {
	case <-e.Server.ReadyNotify():
		fmt.Printf("Embedded etcd is ready on port %d!\n", port)
	case <-time.After(10 * time.Second):
		e.Server.Stop() // trigger a shutdown
		return nil, 0, fmt.Errorf("server took too long to start")
	}

	return e, port, nil
}

func PickAvailableRandomPort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port, err := ln.Addr().(*net.TCPAddr).Port, ln.Close()
	if err != nil {
		return 0, err
	}
	return port, err
}

// StopEmbeddedEtcd stops the embedded etcd server and removes the data directory.
func StopEmbeddedEtcd(e *embed.Etcd) {
	e.Close()
	_ = os.RemoveAll(e.Config().Dir)
	fmt.Println("Embedded etcd server stopped and data directory removed")
}

func createTempDir() (string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "gokube-test-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Return the absolute path of the created directory
	return filepath.Abs(tempDir)
}
