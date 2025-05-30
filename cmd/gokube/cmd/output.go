package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"gokube/pkg/api"

	"gopkg.in/yaml.v2"
)

// TableFormatter is a function type for formatting resources as tables
type TableFormatter func(w io.Writer, data interface{}) error

// output is a generic function that handles JSON, YAML, and table formatting
func output(w io.Writer, data interface{}, format string, tableFormatter TableFormatter) error {
	switch strings.ToLower(format) {
	case "json":
		data, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(data))
	case "yaml":
		data, err := yaml.Marshal(data)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(data))
	default:
		return tableFormatter(w, data)
	}
	return nil
}

// Table formatters for single resources
func formatPodTable(w io.Writer, data interface{}) error {
	pod := data.(*api.Pod)
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tNODE\tIMAGE\tAGE")
	image := ""
	if len(pod.Spec.Containers) > 0 {
		image = pod.Spec.Containers[0].Image
	}
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
		pod.Name,
		pod.Status,
		pod.NodeName,
		image,
		"<unknown>") // Age would require creation timestamp
	tw.Flush()
	return nil
}

func formatNodeTable(w io.Writer, data interface{}) error {
	node := data.(*api.Node)
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tSCHEDULABLE\tPROVIDER-ID\tAGE")
	schedulable := "true"
	if node.Spec.Unschedulable {
		schedulable = "false"
	}
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
		node.Name,
		node.Status,
		schedulable,
		node.Spec.ProviderID,
		"<unknown>") // Age would require creation timestamp
	tw.Flush()
	return nil
}

func formatReplicaSetTable(w io.Writer, data interface{}) error {
	rs := data.(*api.ReplicaSet)
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tDESIRED\tCURRENT\tREADY\tAGE")
	fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%s\n",
		rs.Name,
		rs.Spec.Replicas,
		rs.Status.Replicas,
		rs.Status.ReadyReplicas,
		"<unknown>") // Age would require creation timestamp
	tw.Flush()
	return nil
}

// Table formatters for multiple resources
func formatPodsTable(w io.Writer, data interface{}) error {
	pods := data.([]*api.Pod)
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tNODE\tIMAGE\tAGE")
	for _, pod := range pods {
		image := ""
		if len(pod.Spec.Containers) > 0 {
			image = pod.Spec.Containers[0].Image
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			pod.Name,
			pod.Status,
			pod.NodeName,
			image,
			"<unknown>")
	}
	tw.Flush()
	return nil
}

func formatNodesTable(w io.Writer, data interface{}) error {
	nodes := data.([]*api.Node)
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tSCHEDULABLE\tPROVIDER-ID\tAGE")
	for _, node := range nodes {
		schedulable := "true"
		if node.Spec.Unschedulable {
			schedulable = "false"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			node.Name,
			node.Status,
			schedulable,
			node.Spec.ProviderID,
			"<unknown>")
	}
	tw.Flush()
	return nil
}

func formatReplicaSetsTable(w io.Writer, data interface{}) error {
	replicasets := data.([]*api.ReplicaSet)
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "NAME\tDESIRED\tCURRENT\tREADY\tAGE")
	for _, rs := range replicasets {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%s\n",
			rs.Name,
			rs.Spec.Replicas,
			rs.Status.Replicas,
			rs.Status.ReadyReplicas,
			"<unknown>")
	}
	tw.Flush()
	return nil
}

// Public output functions that use the generic output function
func outputPod(w io.Writer, pod *api.Pod, format string) error {
	return output(w, pod, format, formatPodTable)
}

func outputPods(w io.Writer, pods []*api.Pod, format string) error {
	return output(w, pods, format, formatPodsTable)
}

func outputNode(w io.Writer, node *api.Node, format string) error {
	return output(w, node, format, formatNodeTable)
}

func outputNodes(w io.Writer, nodes []*api.Node, format string) error {
	return output(w, nodes, format, formatNodesTable)
}

func outputReplicaSet(w io.Writer, rs *api.ReplicaSet, format string) error {
	return output(w, rs, format, formatReplicaSetTable)
}

func outputReplicaSets(w io.Writer, replicasets []*api.ReplicaSet, format string) error {
	return output(w, replicasets, format, formatReplicaSetsTable)
}
