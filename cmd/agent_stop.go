package cmd

import (
	"fmt"

	"github.com/fixpanic/fixpanic-cli/internal/connectivity"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/fixpanic/fixpanic-cli/internal/service"
	"github.com/spf13/cobra"
)

// agentStopCmd represents the agent stop command
var agentStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Fixpanic agent",
	Long: `Stop the Fixpanic agent service.
	
This command stops the agent service using systemd if available.`,
	Example: `  # Stop the agent
  fixpanic agent stop`,
	RunE: runAgentStop,
}

func init() {
	agentCmd.AddCommand(agentStopCmd)
}

func runAgentStop(cmd *cobra.Command, args []string) error {
	fmt.Println("Stopping Fixpanic agent...")

	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if connectivity layer is installed
	connectivityManager := connectivity.NewManager(platformInfo)
	if !connectivityManager.IsInstalled() {
		return fmt.Errorf("agent is not installed")
	}

	// Try to use systemd service if available
	if platform.IsSystemdAvailable() {
		serviceManager := service.NewManager(platformInfo)

		// Check current status
		status, err := serviceManager.Status()
		if err != nil {
			fmt.Printf("Warning: could not check service status: %v\n", err)
		} else if status != "active" {
			fmt.Println("ℹ️  Agent service is not running")
			return nil
		}

		// Stop the service
		if err := serviceManager.Stop(); err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}

		fmt.Println("✅ Agent service stopped successfully")
		fmt.Printf("Service: %s\n", platform.GetSystemdServiceName())

		return nil
	}

	// If systemd is not available, we can't stop a running service
	// The user would have had to start it manually in the foreground
	fmt.Println("ℹ️  Systemd not available.")
	fmt.Println("If you started the agent manually, press Ctrl+C in the terminal where it's running.")

	return nil
}
