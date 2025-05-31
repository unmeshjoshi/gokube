package main

import (
	"fmt"
	"os"

	"gokube/cmd/gokube/cmd"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gokube",
		Short: "A CLI tool for managing gokube resources",
		Long: `gokube is a CLI tool for managing gokube cluster resources including pods, nodes, and replicasets.
It provides CRUD operations for all resource types and connects to the gokube API server.

Examples:
  # Get resources
  gokube get pods
  gokube get pod my-pod
  gokube get nodes
  gokube get node my-node
  gokube get replicasets

  # Create resources
  gokube create pod my-pod --image nginx
  gokube create replicaset my-rs --replicas 3 --image nginx

  # Apply resources (create or update)
  gokube apply pod my-pod --image nginx:latest
  gokube apply node my-node --unschedulable
  gokube apply replicaset my-rs --replicas 5

  # Edit resources
  gokube edit pod my-pod
  gokube edit node my-node
  gokube edit replicaset my-rs

  # Delete resources
  gokube delete pod my-pod
  gokube delete replicaset my-rs

  # Scale resources
  gokube scale replicaset my-rs --replicas 10`,
	}

	// Add verb-based subcommands
	rootCmd.AddCommand(cmd.NewGetCommand())
	rootCmd.AddCommand(cmd.NewCreateCommand())
	rootCmd.AddCommand(cmd.NewApplyCommand())
	rootCmd.AddCommand(cmd.NewEditCommand())
	rootCmd.AddCommand(cmd.NewDeleteCommand())
	rootCmd.AddCommand(cmd.NewScaleCommand())
	rootCmd.AddCommand(cmd.NewVersionCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
