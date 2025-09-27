package cmd

import (
	"fmt"

	"github.com/fixpanic/fixpanic-cli/internal/connectivity"
	"github.com/fixpanic/fixpanic-cli/internal/logger"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/fixpanic/fixpanic-cli/internal/process"
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

	// Use cross-platform process manager for direct process execution
	configPath := platformInfo.GetConfigPath()

	fmt.Printf("Starting: %s --config %s\n", binaryPath, configPath)

	// Create process manager for the current platform
	procManager := process.NewProcessManager()

	// Start the agent process
	procInfo, err := procManager.StartProcess(process.ProcessConfig{
		BinaryPath: binaryPath,
		Args:       []string{"--config", configPath},
		Detach:     true,
	})
	if err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	fmt.Println("✅ Agent started successfully in background")
	fmt.Printf("Process PID: %d\n", procInfo.PID)

	return nil
}
