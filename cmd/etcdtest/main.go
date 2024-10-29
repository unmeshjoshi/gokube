package main

import (
	"fmt"
	"log"

	"gokube/pkg/storage"
)

func main() {
	fmt.Println("Starting etcd test application")

	etcdServer, _, err := storage.StartEmbeddedEtcd()
	if err != nil {
		log.Fatalf("Failed to start embedded etcd: %v", err)
	}
	defer storage.StopEmbeddedEtcd(etcdServer)

	fmt.Println("Embedded etcd server is running")

	// Add your main application logic here

	fmt.Println("Etcd test application completed")
}
