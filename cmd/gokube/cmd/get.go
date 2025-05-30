package cmd

import (
	"context"
	"fmt"
	"strings"

	"gokube/pkg/api"
	"gokube/pkg/sdk"

	"github.com/spf13/cobra"
)

var (
	getOutputFormat string
	getNodeName     string
	getUnassigned   bool
)

// NewGetCommand creates the get command with support for different resources
func NewGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [resource] [name]",
		Short: "Get resources",
		Long:  "Get one or many resources by name or get all resources of a specific type",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceType := strings.ToLower(args[0])

			// Handle plural forms
			switch resourceType {
			case "pods":
				resourceType = "pod"
			case "nodes":
				resourceType = "node"
			case "replicasets", "rs":
				resourceType = "replicaset"
			}

			switch resourceType {
			case "pod":
				return handleGetPod(cmd, args)
			case "node":
				return handleGetNode(cmd, args)
			case "replicaset":
				return handleGetReplicaSet(cmd, args)
			default:
				return fmt.Errorf("unsupported resource type: %s", args[0])
			}
		},
	}

	AddGlobalFlags(cmd)
	cmd.Flags().StringVarP(&getOutputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVar(&getNodeName, "node", "", "Filter pods by node name")
	cmd.Flags().BoolVar(&getUnassigned, "unassigned", false, "Show only unassigned pods")

	return cmd
}

func handleGetPod(cmd *cobra.Command, args []string) error {
	client := NewClient()
	ctx := context.Background()

	if len(args) > 1 {
		// Get specific pod
		pod, err := client.Pods().Get(ctx, args[1])
		if err != nil {
			return fmt.Errorf("failed to get pod: %w", err)
		}
		return outputPod(cmd.OutOrStdout(), pod, getOutputFormat)
	} else {
		// List pods
		var pods []*api.Pod
		var err error

		if getUnassigned {
			pods, err = client.Pods().ListUnassigned(ctx)
		} else if getNodeName != "" {
			pods, err = client.Pods().List(ctx, sdk.WithNodeName(getNodeName))
		} else {
			pods, err = client.Pods().List(ctx)
		}

		if err != nil {
			return fmt.Errorf("failed to list pods: %w", err)
		}
		return outputPods(cmd.OutOrStdout(), pods, getOutputFormat)
	}
}

func handleGetNode(cmd *cobra.Command, args []string) error {
	client := NewClient()
	ctx := context.Background()

	if len(args) > 1 {
		// Get specific node
		node, err := client.Nodes().Get(ctx, args[1])
		if err != nil {
			return fmt.Errorf("failed to get node: %w", err)
		}
		return outputNode(cmd.OutOrStdout(), node, getOutputFormat)
	} else {
		// List nodes
		nodes, err := client.Nodes().List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list nodes: %w", err)
		}
		return outputNodes(cmd.OutOrStdout(), nodes, getOutputFormat)
	}
}

func handleGetReplicaSet(cmd *cobra.Command, args []string) error {
	client := NewClient()
	ctx := context.Background()

	if len(args) > 1 {
		// Get specific replicaset
		rs, err := client.ReplicaSets().Get(ctx, args[1])
		if err != nil {
			return fmt.Errorf("failed to get replicaset: %w", err)
		}
		return outputReplicaSet(cmd.OutOrStdout(), rs, getOutputFormat)
	} else {
		// List replicasets
		replicasets, err := client.ReplicaSets().List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list replicasets: %w", err)
		}
		return outputReplicaSets(cmd.OutOrStdout(), replicasets, getOutputFormat)
	}
}
