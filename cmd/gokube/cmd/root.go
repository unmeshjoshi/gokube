package cmd

import (
	"fmt"
	"os"
	"time"

	"gokube/pkg/sdk"

	"github.com/spf13/cobra"
)

const (
	// Default API server URL
	defaultAPIServerURL = "http://localhost:8080"
	// Environment variable name for API server URL
	apiServerEnvVar = "GOKUBE_API_SERVER"
)

var (
	// Global flags
	apiServerURL string
	timeout      time.Duration

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
		BaseURL: apiServerURL,
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
	// API server URL - can be overridden by environment variable or flag
	apiServerDefault := defaultAPIServerURL
	if envAPIServerURL := os.Getenv(apiServerEnvVar); envAPIServerURL != "" {
		apiServerDefault = envAPIServerURL
	}

	cmd.PersistentFlags().StringVar(&apiServerURL, "api-server", apiServerDefault, fmt.Sprintf("gokube API server URL (can be set via %s environment variable)", apiServerEnvVar))
	cmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "request timeout")
}
