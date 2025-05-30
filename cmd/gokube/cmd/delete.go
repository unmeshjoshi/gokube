package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the delete command with support for different resources
func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [resource] [name]",
		Short: "Delete resources",
		Long:  "Delete resources like pods and replicasets",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceType := strings.ToLower(args[0])

			// Handle plural forms
			switch resourceType {
			case "pods":
				resourceType = "pod"
			case "replicasets", "rs":
				resourceType = "replicaset"
			}

			switch resourceType {
			case "pod":
				return handleDeletePod(cmd, args)
			case "replicaset":
				return handleDeleteReplicaSet(cmd, args)
			default:
				return fmt.Errorf("unsupported resource type for delete: %s", args[0])
			}
		},
	}

	AddGlobalFlags(cmd)

	return cmd
}

func handleDeletePod(cmd *cobra.Command, args []string) error {
	client := NewClient()
	ctx := context.Background()

	err := client.Pods().Delete(ctx, args[1])
	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "pod \"%s\" deleted\n", args[1])
	return nil
}

func handleDeleteReplicaSet(cmd *cobra.Command, args []string) error {
	client := NewClient()
	ctx := context.Background()

	err := client.ReplicaSets().Delete(ctx, args[1])
	if err != nil {
		return fmt.Errorf("failed to delete replicaset: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "replicaset \"%s\" deleted\n", args[1])
	return nil
}
