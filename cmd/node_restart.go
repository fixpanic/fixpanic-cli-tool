package cmd

import (
	"fmt"

	"github.com/fixpanic/opssquad-cli-tool/internal/logger"
	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
	"github.com/fixpanic/opssquad-cli-tool/internal/process"
	"github.com/spf13/cobra"
)

// nodeRestartCmd represents the node restart command
var nodeRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart OpsSquad node",
	Long: `Restart the OpsSquad node service.

This command stops the node if it's running and then starts it again.
It's equivalent to running 'opssquad node stop' followed by 'opssquad node start'.`,
	Example: `  # Restart the node
  opssquad node restart`,
	RunE: runNodeRestart,
}

func init() {
	nodeCmd.AddCommand(nodeRestartCmd)
}

func runNodeRestart(cmd *cobra.Command, args []string) error {
	logger.Header("Restarting OpsSquad Node")

	// Stop the node first
	logger.Step(1, "Stopping node")
	if err := stopNode(); err != nil {
		// If stop fails, continue with start (node might not be running)
		logger.Warning("Stop failed: %v", err)
		logger.Info("Continuing with start...")
	} else {
		logger.Success("Node stopped successfully")
	}

	// Wait a moment for cleanup
	logger.Progress("Waiting for cleanup...")

	// Start the node
	logger.Step(2, "Starting node")
	if err := startNode(); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	logger.Separator()
	logger.Success("OpsSquad node restarted successfully!")

	return nil
}

// stopNode stops all running node processes
func stopNode() error {
	// Get all running node processes
	pids, err := getAllNodeProcessPIDs()
	if err != nil {
		return fmt.Errorf("failed to check node status: %w", err)
	}

	if len(pids) == 0 {
		logger.Info("OpsSquad Node is not running")
		return nil
	}

	// Create process manager for the current platform
	procManager := process.NewProcessManager()

	// Stop all node processes
	stoppedCount := 0
	for _, pid := range pids {
		logger.Progress("Stopping OpsSquad Node (PID: %d)...", pid)
		if err := procManager.StopProcess(pid); err != nil {
			logger.Warning("Failed to stop process %d: %v", pid, err)
		} else {
			stoppedCount++
		}
	}

	if stoppedCount == 0 {
		return fmt.Errorf("failed to stop any node processes")
	}

	if stoppedCount == 1 {
		logger.Success("OpsSquad Node stopped successfully")
	} else {
		logger.Success("OpsSquad Node stopped successfully (%d processes stopped)", stoppedCount)
	}
	return nil
}

// startNode starts the node
func startNode() error {
	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Validate node installation
	connectivityManager, err := validateNodeInstall(platformInfo)
	if err != nil {
		return err
	}

	// Clean up old nodes
	if err := cleanUpOldNodes(); err != nil {
		return err
	}

	// Start the node service
	return startNodeService(platformInfo, connectivityManager)
}