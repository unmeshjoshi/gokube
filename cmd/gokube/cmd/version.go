package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// These will be set by build flags
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// NewVersionCommand creates the version command
func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  "Print the version, git commit, and build date information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("gokube CLI version %s\n", Version)
			fmt.Printf("Git commit: %s\n", GitCommit)
			fmt.Printf("Build date: %s\n", BuildDate)
			return nil
		},
	}

	return cmd
}
