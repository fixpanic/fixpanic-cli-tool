package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fixpanic/fixpanic-cli/internal/connectivity"
	"github.com/fixpanic/fixpanic-cli/internal/logger"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/fixpanic/fixpanic-cli/internal/service"
	"github.com/spf13/cobra"
)

// agentStartCmd represents the agent start command
var agentStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Fixpanic agent",
	Long: `Start the Fixpanic agent service.
	
This command starts the agent service using systemd if available, or runs the
connectivity layer binary directly if systemd is not available.`,
	Example: `  # Start the agent
  fixpanic agent start`,
	RunE: runAgentStart,
}

func init() {
	agentCmd.AddCommand(agentStartCmd)
}

func runAgentStart(cmd *cobra.Command, args []string) error {
	logger.Header("Starting FixPanic Agent")

	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if agent is installed and ensure it's the latest version
	logger.Step(1, "Checking agent installation and updates")
	connectivityManager := connectivity.NewManager(platformInfo)

	if !connectivityManager.IsFixPanicAgentInstalled() {
		return fmt.Errorf("FixPanic Agent not installed. Run 'fixpanic agent install' first")
	}

	// Ensure we have the latest version before starting
	if err := connectivityManager.EnsureLatestAgent(); err != nil {
		logger.Warning("Failed to check for agent updates: %v", err)
		logger.Info("Continuing with existing agent binary")
	}

	binaryPath := platformInfo.GetFixPanicAgentBinaryPath()

	// Try to use systemd service if available
	logger.Step(2, "Starting agent service")
	if platform.IsSystemdAvailable() {
		serviceManager := service.NewManager(platformInfo)

		// Check current status
		logger.Progress("Checking service status")
		status, err := serviceManager.Status()
		if err != nil {
			fmt.Printf("Warning: could not check service status: %v\n", err)
		} else if status == "active" {
			fmt.Println("✅ Agent service is already running")
			return nil
		}

		// Start the service
		if err := serviceManager.Start(); err != nil {
			return fmt.Errorf("failed to start service: %w", err)
		}

		fmt.Println("✅ Agent service started successfully")
		fmt.Printf("Service: %s\n", platform.GetSystemdServiceName())

		// Show how to check status
		fmt.Println("\nYou can check the status with:")
		fmt.Printf("  sudo systemctl status %s\n", platform.GetSystemdServiceName())

		return nil
	}

	// Fallback: run the binary directly (not recommended for production)
	fmt.Println("⚠️  Systemd not available. Starting agent directly...")
	fmt.Println("Note: This is not recommended for production use.")
	fmt.Println("The agent will not restart automatically if it crashes.")

	configPath := platformInfo.GetConfigPath()

	fmt.Printf("Starting: %s --config %s\n", binaryPath, configPath)
	fmt.Println("Press Ctrl+C to stop the agent")

	// Execute the binary directly
	execCmd := exec.Command(binaryPath, "--config", configPath)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	return nil
}
