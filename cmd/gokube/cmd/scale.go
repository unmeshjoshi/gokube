package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	scaleReplicas     int32
	scaleOutputFormat string
)

// NewScaleCommand creates the scale command for replicasets
func NewScaleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scale [resource] [name]",
		Short: "Scale a resource",
		Long:  "Scale a replicaset to a specified number of replicas",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceType := strings.ToLower(args[0])

			// Handle plural forms
			switch resourceType {
			case "replicasets", "rs":
				resourceType = "replicaset"
			}

			switch resourceType {
			case "replicaset":
				return handleScaleReplicaSet(cmd, args)
			default:
				return fmt.Errorf("unsupported resource type for scale: %s", args[0])
			}
		},
	}

	AddGlobalFlags(cmd)
	cmd.Flags().Int32Var(&scaleReplicas, "replicas", 1, "Number of replicas to scale to")
	cmd.Flags().StringVarP(&scaleOutputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.MarkFlagRequired("replicas")

	return cmd
}

func handleScaleReplicaSet(cmd *cobra.Command, args []string) error {
	client := NewClient()
	ctx := context.Background()

	// Get the existing replicaset
	existingRS, err := client.ReplicaSets().Get(ctx, args[1])
	if err != nil {
		return fmt.Errorf("failed to get replicaset: %w", err)
	}

	// Update replica count
	existingRS.Spec.Replicas = scaleReplicas

	result, err := client.ReplicaSets().Update(ctx, existingRS)
	if err != nil {
		return fmt.Errorf("failed to scale replicaset: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "replicaset \"%s\" scaled to %d replicas\n", args[1], scaleReplicas)
	return outputReplicaSet(cmd.OutOrStdout(), result, scaleOutputFormat)
}
