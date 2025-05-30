package cmd

import (
	"time"

	"gokube/pkg/sdk"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	serverURL string
	timeout   time.Duration

	// For testing - allows injection of mock client
	testClient sdk.ClientInterface
)

// NewClient creates a new SDK client with the configured parameters
func NewClient() sdk.ClientInterface {
	// If a test client is set, use it
	if testClient != nil {
		return testClient
	}

	config := sdk.Config{
		BaseURL: serverURL,
		Timeout: timeout,
	}
	return sdk.NewClient(config)
}

// SetTestClient sets a mock client for testing (internal use only)
func SetTestClient(client sdk.ClientInterface) {
	testClient = client
}

// ResetTestClient resets the test client (internal use only)
func ResetTestClient() {
	testClient = nil
}

// AddGlobalFlags adds global flags to a command
func AddGlobalFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "gokube API server URL")
	cmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "request timeout")
}
