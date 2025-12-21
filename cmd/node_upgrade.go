package cmd

import (
	"fmt"

	"github.com/fixpanic/opssquad-cli-tool/internal/connectivity"
	"github.com/fixpanic/opssquad-cli-tool/internal/logger"
	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
	"github.com/fixpanic/opssquad-cli-tool/internal/process"
	"github.com/spf13/cobra"
)

var forceNodeUpgrade bool

// nodeUpgradeCmd represents the node upgrade command
var nodeUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade OpsSquad node to latest version",
	Long: `Upgrade the OpsSquad node binary to the latest version.

This command downloads and installs the latest version of the connectivity
layer binary, ensuring your node has the latest features and security updates.`,
	Example: `  # Upgrade node to latest version
  opssquad node upgrade

  # Force upgrade even if already on latest version
  opssquad node upgrade --force`,
	RunE: runNodeUpgrade,
}

func init() {
	nodeCmd.AddCommand(nodeUpgradeCmd)

	// Add flags
	nodeUpgradeCmd.Flags().BoolVar(&forceNodeUpgrade, "force", false, "Force upgrade even if already on latest version")
}

func runNodeUpgrade(cmd *cobra.Command, args []string) error {
	logger.Header("Upgrading OpsSquad Node")

	// Get platform information
	logger.Step(1, "Detecting platform and configuration")
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if OpsSquad Node is installed
	logger.Step(2, "Checking node installation")
	connectivityManager := connectivity.NewManager(platformInfo)
	if !connectivityManager.IsOpsSquadNodeInstalled() {
		return fmt.Errorf("OpsSquad Node is not installed. Run 'opssquad node install' first")
	}

	// Get current version
	logger.Progress("Checking current node version")
	currentVersion, err := connectivityManager.GetOpsSquadNodeVersion()
	if err != nil {
		logger.Warning("Could not determine current version: %v", err)
		currentVersion = "unknown"
	} else {
		logger.KeyValue("Current version", currentVersion)
	}

	// Check if node is running and stop it before upgrade
	logger.Step(3, "Stopping node for upgrade")
	nodeWasRunning := false
	pids, err := getAllNodeProcessPIDs()
	if err != nil {
		logger.Warning("Failed to check node status: %v", err)
	} else if len(pids) > 0 {
		// Node is running, stop it to allow binary replacement
		logger.Progress("Stopping running node to allow binary replacement")
		procManager := process.NewProcessManager()
		stoppedCount := 0
		for _, pid := range pids {
			logger.Progress("Stopping node process (PID: %d)", pid)
			if err := procManager.StopProcess(pid); err != nil {
				logger.Warning("Failed to stop process %d: %v", pid, err)
			} else {
				stoppedCount++
			}
		}
		if stoppedCount > 0 {
			nodeWasRunning = true
			logger.Success("Node stopped successfully (%d process(es) stopped)", stoppedCount)
		} else {
			logger.Warning("Failed to stop node, attempting upgrade anyway...")
		}
	} else {
		logger.Info("Node is not running, proceeding with upgrade")
	}

	// Upgrade node binary
	logger.Step(4, "Upgrading node binary")
	if err := connectivityManager.EnsureLatestNode(); err != nil {
		return fmt.Errorf("failed to upgrade node binary: %w", err)
	}

	// Get new version
	logger.Progress("Verifying upgrade")
	newVersion, err := connectivityManager.GetOpsSquadNodeVersion()
	if err != nil {
		logger.Warning("Could not determine new version: %v", err)
		newVersion = "unknown"
	} else {
		logger.KeyValue("New version", newVersion)
	}

	// Check if upgrade was needed
	if !forceNodeUpgrade && currentVersion == newVersion && currentVersion != "unknown" {
		logger.Success("Node was already on the latest version")
	} else {
		logger.Success("Node upgraded successfully!")
		if currentVersion != "unknown" && newVersion != "unknown" {
			logger.Info("Upgraded: %s → %s", currentVersion, newVersion)
		}
	}

	// Restart node if it was running before upgrade
	if nodeWasRunning {
		logger.Step(5, "Restarting node")
		logger.Progress("Starting node with new version")

		// Use node start command to restart
		if err := nodeStartCmd.RunE(cmd, []string{}); err != nil {
			logger.Warning("Failed to restart node: %v", err)
			logger.Info("You can start the node manually with: opssquad node start")
		} else {
			logger.Success("Node restarted successfully with new version")
		}
	}

	logger.Separator()
	logger.KeyValue("Binary location", platformInfo.GetOpsSquadNodeBinaryPath())

	if !nodeWasRunning {
		logger.Info("Node was not running before upgrade")
		logger.Info("You can start the node with: opssquad node start")
	}

	return nil
}