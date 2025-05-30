package cmd

import (
	"context"
	"fmt"
	"strings"

	"gokube/pkg/api"

	"github.com/spf13/cobra"
)

var (
	createOutputFormat string
	createImage        string
	createLabels       []string
	createReplicas     int32
)

// NewCreateCommand creates the create command with support for different resources
func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [resource] [name]",
		Short: "Create resources",
		Long:  "Create resources like pods and replicasets",
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
				return handleCreatePod(cmd, args)
			case "replicaset":
				return handleCreateReplicaSet(cmd, args)
			default:
				return fmt.Errorf("unsupported resource type for create: %s", args[0])
			}
		},
	}

	AddGlobalFlags(cmd)
	cmd.Flags().StringVarP(&createOutputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVar(&createImage, "image", "nginx", "Container image")
	cmd.Flags().StringSliceVar(&createLabels, "labels", []string{}, "Labels in key=value format")
	cmd.Flags().Int32Var(&createReplicas, "replicas", 1, "Number of replicas for replicaset")

	return cmd
}

func handleCreatePod(cmd *cobra.Command, args []string) error {
	client := NewClient()
	ctx := context.Background()

	pod := &api.Pod{
		ObjectMeta: api.ObjectMeta{
			Name: args[1],
		},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{
					Name:  "main",
					Image: createImage,
				},
			},
		},
	}

	result, err := client.Pods().Create(ctx, pod)
	if err != nil {
		return fmt.Errorf("failed to create pod: %w", err)
	}

	return outputPod(cmd.OutOrStdout(), result, createOutputFormat)
}

func handleCreateReplicaSet(cmd *cobra.Command, args []string) error {
	client := NewClient()
	ctx := context.Background()

	replicaset := &api.ReplicaSet{
		ObjectMeta: api.ObjectMeta{
			Name: args[1],
		},
		Spec: api.ReplicaSetSpec{
			Replicas: createReplicas,
			Template: api.PodTemplateSpec{
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  "main",
							Image: createImage,
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

	return outputReplicaSet(cmd.OutOrStdout(), result, createOutputFormat)
}
