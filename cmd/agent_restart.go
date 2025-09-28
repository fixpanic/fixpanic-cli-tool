package cmd

import (
	"fmt"

	"github.com/fixpanic/fixpanic-cli/internal/logger"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/fixpanic/fixpanic-cli/internal/process"
	"github.com/spf13/cobra"
)

// agentRestartCmd represents the agent restart command
var agentRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart Fixpanic agent",
	Long: `Restart the Fixpanic agent service.

This command stops the agent if it's running and then starts it again.
It's equivalent to running 'fixpanic agent stop' followed by 'fixpanic agent start'.`,
	Example: `  # Restart the agent
  fixpanic agent restart`,
	RunE: runAgentRestart,
}

func init() {
	agentCmd.AddCommand(agentRestartCmd)
}

func runAgentRestart(cmd *cobra.Command, args []string) error {
	logger.Header("Restarting FixPanic Agent")

	// Stop the agent first
	logger.Step(1, "Stopping agent")
	if err := stopAgent(); err != nil {
		// If stop fails, continue with start (agent might not be running)
		logger.Warning("Stop failed: %v", err)
		logger.Info("Continuing with start...")
	} else {
		logger.Success("Agent stopped successfully")
	}

	// Wait a moment for cleanup
	logger.Progress("Waiting for cleanup...")

	// Start the agent
	logger.Step(2, "Starting agent")
	if err := startAgent(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	logger.Separator()
	logger.Success("FixPanic agent restarted successfully!")

	return nil
}

// stopAgent stops all running agent processes
func stopAgent() error {
	// Get all running agent processes
	pids, err := getAllAgentProcessPIDs()
	if err != nil {
		return fmt.Errorf("failed to check agent status: %w", err)
	}

	if len(pids) == 0 {
		logger.Info("FixPanic Agent is not running")
		return nil
	}

	// Create process manager for the current platform
	procManager := process.NewProcessManager()

	// Stop all agent processes
	stoppedCount := 0
	for _, pid := range pids {
		logger.Progress("Stopping FixPanic Agent (PID: %d)...", pid)
		if err := procManager.StopProcess(pid); err != nil {
			logger.Warning("Failed to stop process %d: %v", pid, err)
		} else {
			stoppedCount++
		}
	}

	if stoppedCount == 0 {
		return fmt.Errorf("failed to stop any agent processes")
	}

	if stoppedCount == 1 {
		logger.Success("FixPanic Agent stopped successfully")
	} else {
		logger.Success("FixPanic Agent stopped successfully (%d processes stopped)", stoppedCount)
	}
	return nil
}

// startAgent starts the agent
func startAgent() error {
	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Validate agent installation
	connectivityManager, err := validateAgentInstall(platformInfo)
	if err != nil {
		return err
	}

	// Clean up old agents
	if err := cleanUpOldAgents(); err != nil {
		return err
	}

	// Start the agent service
	return startAgentService(platformInfo, connectivityManager)
}