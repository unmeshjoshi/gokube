package cmd

import (
	"context"
	"fmt"
	"strings"

	"gokube/pkg/api"

	"github.com/spf13/cobra"
)

var (
	applyOutputFormat  string
	applyUnschedulable bool
	applyProviderID    string
	applyImage         string
	applyNodeName      string
	applyReplicas      int32
)

// NewApplyCommand creates the apply command with support for different resources
func NewApplyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply [resource] [name]",
		Short: "Apply (create or update) resources",
		Long:  "Apply resources like pods, nodes, and replicasets. Creates the resource if it doesn't exist, updates it if it does (upsert behavior).",
		Args:  cobra.ExactArgs(2),
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
				return handleApplyPod(cmd, args)
			case "node":
				return handleApplyNode(cmd, args)
			case "replicaset":
				return handleApplyReplicaSet(cmd, args)
			default:
				return fmt.Errorf("unsupported resource type for apply: %s", args[0])
			}
		},
	}

	AddGlobalFlags(cmd)
	cmd.Flags().StringVarP(&applyOutputFormat, "output", "o", "table", "Output format (table, json, yaml)")

	// Node flags
	cmd.Flags().BoolVar(&applyUnschedulable, "unschedulable", false, "Mark node as unschedulable")
	cmd.Flags().StringVar(&applyProviderID, "provider-id", "", "Cloud provider ID")

	// Pod flags
	cmd.Flags().StringVar(&applyImage, "image", "nginx", "Container image")
	cmd.Flags().StringVar(&applyNodeName, "node", "", "Assign to node")

	// ReplicaSet flags
	cmd.Flags().Int32Var(&applyReplicas, "replicas", 1, "Number of replicas")

	return cmd
}

func handleApplyPod(cmd *cobra.Command, args []string) error {
	client := NewClient()
	ctx := context.Background()

	// Try to get the existing pod first
	existingPod, err := client.Pods().Get(ctx, args[1])
	if err != nil {
		// If pod doesn't exist, create it
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			pod := &api.Pod{
				ObjectMeta: api.ObjectMeta{
					Name: args[1],
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  "main",
							Image: applyImage,
						},
					},
				},
			}

			if applyNodeName != "" {
				pod.NodeName = applyNodeName
			}

			result, err := client.Pods().Create(ctx, pod)
			if err != nil {
				return fmt.Errorf("failed to create pod: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "pod \"%s\" created\n", args[1])
			return outputPod(cmd.OutOrStdout(), result, applyOutputFormat)
		}
		return fmt.Errorf("failed to get existing pod: %w", err)
	}

	// Pod exists, update it
	if cmd.Flags().Changed("image") {
		if len(existingPod.Spec.Containers) > 0 {
			existingPod.Spec.Containers[0].Image = applyImage
		}
	}

	if cmd.Flags().Changed("node") {
		existingPod.NodeName = applyNodeName
	}

	result, err := client.Pods().Update(ctx, existingPod)
	if err != nil {
		return fmt.Errorf("failed to update pod: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "pod \"%s\" configured\n", args[1])
	return outputPod(cmd.OutOrStdout(), result, applyOutputFormat)
}

func handleApplyNode(cmd *cobra.Command, args []string) error {
	client := NewClient()
	ctx := context.Background()

	// Try to get the existing node first
	existingNode, err := client.Nodes().Get(ctx, args[1])
	if err != nil {
		// Node operations don't support create, so we can only update
		return fmt.Errorf("failed to get node (nodes cannot be created via CLI): %w", err)
	}

	// Node exists, update it
	if cmd.Flags().Changed("unschedulable") {
		existingNode.Spec.Unschedulable = applyUnschedulable
	}

	if cmd.Flags().Changed("provider-id") {
		existingNode.Spec.ProviderID = applyProviderID
	}

	result, err := client.Nodes().Update(ctx, existingNode)
	if err != nil {
		return fmt.Errorf("failed to update node: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "node \"%s\" configured\n", args[1])
	return outputNode(cmd.OutOrStdout(), result, applyOutputFormat)
}

func handleApplyReplicaSet(cmd *cobra.Command, args []string) error {
	client := NewClient()
	ctx := context.Background()

	// Try to get the existing replicaset first
	existingRS, err := client.ReplicaSets().Get(ctx, args[1])
	if err != nil {
		// If replicaset doesn't exist, create it
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			replicaset := &api.ReplicaSet{
				ObjectMeta: api.ObjectMeta{
					Name: args[1],
				},
				Spec: api.ReplicaSetSpec{
					Replicas: applyReplicas,
					Template: api.PodTemplateSpec{
						Spec: api.PodSpec{
							Containers: []api.Container{
								{
									Name:  "main",
									Image: applyImage,
								},
							},
						},
					},
				},
			}

			result, err := client.ReplicaSets().Create(ctx, replicaset)
			if err != nil {
				return fmt.Errorf("failed to create replicaset: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "replicaset \"%s\" created\n", args[1])
			return outputReplicaSet(cmd.OutOrStdout(), result, applyOutputFormat)
		}
		return fmt.Errorf("failed to get existing replicaset: %w", err)
	}

	// ReplicaSet exists, update it
	if cmd.Flags().Changed("replicas") {
		existingRS.Spec.Replicas = applyReplicas
	}

	if cmd.Flags().Changed("image") {
		if len(existingRS.Spec.Template.Spec.Containers) > 0 {
			existingRS.Spec.Template.Spec.Containers[0].Image = applyImage
		}
	}

	result, err := client.ReplicaSets().Update(ctx, existingRS)
	if err != nil {
		return fmt.Errorf("failed to update replicaset: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "replicaset \"%s\" configured\n", args[1])
	return outputReplicaSet(cmd.OutOrStdout(), result, applyOutputFormat)
}
