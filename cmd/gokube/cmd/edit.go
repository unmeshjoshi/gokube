package cmd

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gokube/pkg/api"
	"gokube/pkg/sdk"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	editOutputFormat string
)

// NewEditCommand creates the edit command
func NewEditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [resource] [name]",
		Short: "Edit a resource in the default editor",
		Long: `Edit a resource in your default editor and apply changes automatically.

The editor will open with the current resource configuration. Save and exit to apply changes.
Exit without saving to abort the edit operation.

Supported resources:
- pod, pods
- node, nodes
- replicaset, replicasets, rs

Examples:
  # Edit a pod
  gokube edit pod my-pod

  # Edit a node
  gokube edit node my-node

  # Edit a replicaset
  gokube edit replicaset my-rs
  gokube edit rs my-rs`,
		Args: cobra.ExactArgs(2),
		RunE: runEdit,
	}

	AddGlobalFlags(cmd)
	cmd.Flags().StringVarP(&editOutputFormat, "output", "o", "yaml", "Output format (only yaml supported for editing)")

	return cmd
}

func normalizeResourceType(resourceType string) string {
	resourceType = strings.ToLower(resourceType)

	// Handle plural forms and aliases
	switch resourceType {
	case "pods":
		return "pod"
	case "nodes":
		return "node"
	case "replicasets", "rs":
		return "replicaset"
	default:
		return resourceType
	}
}

func runEdit(cmd *cobra.Command, args []string) error {
	resourceType := normalizeResourceType(args[0])
	resourceName := args[1]

	client := NewClient()
	ctx := context.Background()

	// Get the current resource
	var currentData []byte
	var err error

	switch resourceType {
	case "pod":
		var pod *api.Pod
		pod, err = client.Pods().Get(ctx, resourceName)
		if err != nil {
			return fmt.Errorf("failed to get pod %s: %w", resourceName, err)
		}
		currentData, err = yaml.Marshal(pod)
	case "node":
		var node *api.Node
		node, err = client.Nodes().Get(ctx, resourceName)
		if err != nil {
			return fmt.Errorf("failed to get node %s: %w", resourceName, err)
		}
		currentData, err = yaml.Marshal(node)
	case "replicaset":
		var rs *api.ReplicaSet
		rs, err = client.ReplicaSets().Get(ctx, resourceName)
		if err != nil {
			return fmt.Errorf("failed to get replicaset %s: %w", resourceName, err)
		}
		currentData, err = yaml.Marshal(rs)
	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal resource to YAML: %w", err)
	}

	// Start the edit loop
	return editLoop(cmd, client, ctx, resourceType, resourceName, currentData, "")
}

func editLoop(cmd *cobra.Command, client sdk.ClientInterface, ctx context.Context, resourceType, resourceName string, data []byte, errorMsg string) error {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("gokube-edit-%s-%s-*.yaml", resourceType, resourceName))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Add error message as comments if present
	content := data
	if errorMsg != "" {
		errorLines := strings.Split(errorMsg, "\n")
		var commentedError []string
		commentedError = append(commentedError, "# Error applying changes:")
		for _, line := range errorLines {
			if strings.TrimSpace(line) != "" {
				commentedError = append(commentedError, "# "+line)
			}
		}
		commentedError = append(commentedError, "# Please fix the error and save, or exit without saving to abort")
		commentedError = append(commentedError, "")

		errorComment := strings.Join(commentedError, "\n")
		content = append([]byte(errorComment), data...)
	}

	// Calculate original hash for change detection
	originalHash := sha256.Sum256(data)

	// Write content to temp file
	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}
	tmpFile.Close()

	// Open editor
	editor := getEditor()
	editorCmd := exec.Command(editor, tmpFile.Name())
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	// Read the edited content
	editedContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	// Parse the YAML to validate it and extract clean content
	// The YAML parser automatically removes all comments
	var tempResource map[string]interface{}
	if err := yaml.Unmarshal(editedContent, &tempResource); err != nil {
		// If there's a YAML error, start the edit loop again with the error message
		return editLoop(cmd, client, ctx, resourceType, resourceName, editedContent, err.Error())
	}

	// Marshal back to clean YAML for comparison
	cleanedContent, err := yaml.Marshal(tempResource)
	if err != nil {
		return fmt.Errorf("failed to marshal cleaned content: %w", err)
	}

	// Check if content was actually changed
	newHash := sha256.Sum256(cleanedContent)
	if newHash == originalHash {
		fmt.Fprintln(cmd.OutOrStdout(), "Edit cancelled, no changes made.")
		return nil
	}

	// Try to apply the changes using the original edited content
	// (not the cleaned content, as we want to preserve the original structure)
	if err := applyEditedResource(client, ctx, resourceType, resourceName, editedContent); err != nil {
		// If there's an error, start the edit loop again with the error message
		return editLoop(cmd, client, ctx, resourceType, resourceName, editedContent, err.Error())
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s/%s edited\n", resourceType, resourceName)
	return nil
}

func getEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	// Default to vim
	return "vim"
}

// updateResource is a generic helper that handles the common pattern of:
// 1. Unmarshal YAML into resource
// 2. Validate resource name hasn't changed
// 3. Call the update function
func updateResource(data []byte, expectedName string, resource interface{}, updateFn func() error) error {
	// Unmarshal YAML into the resource
	if err := yaml.Unmarshal(data, resource); err != nil {
		return fmt.Errorf("invalid YAML format: %w", err)
	}

	// Get the name based on resource type
	var actualName string
	var resourceTypeName string

	switch r := resource.(type) {
	case *api.Pod:
		actualName = r.Name
		resourceTypeName = "pod"
	case *api.Node:
		actualName = r.Name
		resourceTypeName = "node"
	case *api.ReplicaSet:
		actualName = r.Name
		resourceTypeName = "replicaset"
	default:
		return fmt.Errorf("unsupported resource type: %T", resource)
	}

	// Ensure the name matches
	if actualName != expectedName {
		return fmt.Errorf("resource name cannot be changed (expected: %s, got: %s)", expectedName, actualName)
	}

	// Call the specific update function
	if err := updateFn(); err != nil {
		return fmt.Errorf("failed to update %s: %w", resourceTypeName, err)
	}

	return nil
}

func applyEditedResource(client sdk.ClientInterface, ctx context.Context, resourceType, resourceName string, data []byte) error {
	switch resourceType {
	case "pod":
		var pod api.Pod
		if err := updateResource(data, resourceName, &pod, func() error {
			_, err := client.Pods().Update(ctx, &pod)
			return err
		}); err != nil {
			return err
		}

	case "node":
		var node api.Node
		if err := updateResource(data, resourceName, &node, func() error {
			_, err := client.Nodes().Update(ctx, &node)
			return err
		}); err != nil {
			return err
		}

	case "replicaset":
		var rs api.ReplicaSet
		if err := updateResource(data, resourceName, &rs, func() error {
			_, err := client.ReplicaSets().Update(ctx, &rs)
			return err
		}); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	return nil
}
