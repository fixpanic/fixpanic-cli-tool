package cmd

import (
	"fmt"

	"github.com/fixpanic/fixpanic-cli/internal/connectivity"
	"github.com/fixpanic/fixpanic-cli/internal/logger"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/fixpanic/fixpanic-cli/internal/process"
	"github.com/spf13/cobra"
)

var forceAgentUpgrade bool

// agentUpgradeCmd represents the agent upgrade command
var agentUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Fixpanic agent to latest version",
	Long: `Upgrade the Fixpanic agent binary to the latest version.

This command downloads and installs the latest version of the connectivity
layer binary, ensuring your agent has the latest features and security updates.`,
	Example: `  # Upgrade agent to latest version
  fixpanic agent upgrade

  # Force upgrade even if already on latest version
  fixpanic agent upgrade --force`,
	RunE: runAgentUpgrade,
}

func init() {
	agentCmd.AddCommand(agentUpgradeCmd)

	// Add flags
	agentUpgradeCmd.Flags().BoolVar(&forceAgentUpgrade, "force", false, "Force upgrade even if already on latest version")
}

func runAgentUpgrade(cmd *cobra.Command, args []string) error {
	logger.Header("Upgrading FixPanic Agent")

	// Get platform information
	logger.Step(1, "Detecting platform and configuration")
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if FixPanic Agent is installed
	logger.Step(2, "Checking agent installation")
	connectivityManager := connectivity.NewManager(platformInfo)
	if !connectivityManager.IsFixPanicAgentInstalled() {
		return fmt.Errorf("FixPanic Agent is not installed. Run 'fixpanic agent install' first")
	}

	// Get current version
	logger.Progress("Checking current agent version")
	currentVersion, err := connectivityManager.GetFixPanicAgentVersion()
	if err != nil {
		logger.Warning("Could not determine current version: %v", err)
		currentVersion = "unknown"
	} else {
		logger.KeyValue("Current version", currentVersion)
	}

	// Check if agent is running and stop it before upgrade
	logger.Step(3, "Stopping agent for upgrade")
	agentWasRunning := false
	pids, err := getAllAgentProcessPIDs()
	if err != nil {
		logger.Warning("Failed to check agent status: %v", err)
	} else if len(pids) > 0 {
		// Agent is running, stop it to allow binary replacement
		logger.Progress("Stopping running agent to allow binary replacement")
		procManager := process.NewProcessManager()
		stoppedCount := 0
		for _, pid := range pids {
			logger.Progress("Stopping agent process (PID: %d)", pid)
			if err := procManager.StopProcess(pid); err != nil {
				logger.Warning("Failed to stop process %d: %v", pid, err)
			} else {
				stoppedCount++
			}
		}
		if stoppedCount > 0 {
			agentWasRunning = true
			logger.Success("Agent stopped successfully (%d process(es) stopped)", stoppedCount)
		} else {
			logger.Warning("Failed to stop agent, attempting upgrade anyway...")
		}
	} else {
		logger.Info("Agent is not running, proceeding with upgrade")
	}

	// Upgrade agent binary
	logger.Step(4, "Upgrading agent binary")
	if err := connectivityManager.EnsureLatestAgent(); err != nil {
		return fmt.Errorf("failed to upgrade agent binary: %w", err)
	}

	// Get new version
	logger.Progress("Verifying upgrade")
	newVersion, err := connectivityManager.GetFixPanicAgentVersion()
	if err != nil {
		logger.Warning("Could not determine new version: %v", err)
		newVersion = "unknown"
	} else {
		logger.KeyValue("New version", newVersion)
	}

	// Check if upgrade was needed
	if !forceAgentUpgrade && currentVersion == newVersion && currentVersion != "unknown" {
		logger.Success("Agent was already on the latest version")
	} else {
		logger.Success("Agent upgraded successfully!")
		if currentVersion != "unknown" && newVersion != "unknown" {
			logger.Info("Upgraded: %s → %s", currentVersion, newVersion)
		}
	}

	// Restart agent if it was running before upgrade
	if agentWasRunning {
		logger.Step(5, "Restarting agent")
		logger.Progress("Starting agent with new version")

		// Use agent start command to restart
		if err := agentStartCmd.RunE(cmd, []string{}); err != nil {
			logger.Warning("Failed to restart agent: %v", err)
			logger.Info("You can start the agent manually with: fixpanic agent start")
		} else {
			logger.Success("Agent restarted successfully with new version")
		}
	}

	logger.Separator()
	logger.KeyValue("Binary location", platformInfo.GetFixPanicAgentBinaryPath())

	if !agentWasRunning {
		logger.Info("Agent was not running before upgrade")
		logger.Info("You can start the agent with: fixpanic agent start")
	}

	return nil
}