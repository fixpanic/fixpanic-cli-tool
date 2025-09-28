package cmd

import (
	"fmt"

	"github.com/fixpanic/fixpanic-cli/internal/connectivity"
	"github.com/fixpanic/fixpanic-cli/internal/logger"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
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

	// Upgrade agent binary
	logger.Step(3, "Upgrading agent binary")
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

	logger.Separator()
	logger.KeyValue("Binary location", platformInfo.GetFixPanicAgentBinaryPath())
	logger.Info("You can restart the agent with: fixpanic agent restart")

	return nil
}