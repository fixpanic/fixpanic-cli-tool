package cmd

import (
	"fmt"

	"github.com/fixpanic/fixpanic-cli/internal/config"
	"github.com/fixpanic/fixpanic-cli/internal/connectivity"
	"github.com/fixpanic/fixpanic-cli/internal/logger"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/fixpanic/fixpanic-cli/internal/service"
	"github.com/spf13/cobra"
)

var (
	agentID      string
	agentAPIKey  string
	forceInstall bool
)

// agentInstallCmd represents the agent install command
var agentInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Fixpanic agent",
	Long: `Install the Fixpanic agent on your server.

This command downloads and installs the connectivity layer binary, creates the
necessary configuration files, and sets up the systemd service for automatic
startup.`,
	Example: `  # Install with agent credentials
	 fixpanic agent install --agent-id="agent_123" --api-key="fp_abc123xyz"

	 # Force reinstall
	 fixpanic agent install --agent-id="agent_123" --api-key="fp_abc123xyz" --force`,
	RunE: runAgentInstall,
}

func init() {
	agentCmd.AddCommand(agentInstallCmd)

	// Add flags
	agentInstallCmd.Flags().StringVar(&agentID, "agent-id", "", "Agent ID from Fixpanic dashboard (required)")
	agentInstallCmd.Flags().StringVar(&agentAPIKey, "api-key", "", "Agent API key from Fixpanic dashboard (required)")
	agentInstallCmd.Flags().BoolVar(&forceInstall, "force", false, "Force reinstall even if agent is already installed")

	// Mark required flags
	agentInstallCmd.MarkFlagRequired("agent-id")
	agentInstallCmd.MarkFlagRequired("api-key")
}

func runAgentInstall(cmd *cobra.Command, args []string) error {
	logger.Header("Installing Fixpanic Agent")

	// Get platform information
	logger.Step(1, "Detecting platform and configuration")
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if running as root for system-wide installation
	if !platformInfo.IsRoot {
		logger.Warning("Running as non-root user. Agent will be installed in user directories.")
		logger.KeyValue("Binary location", platformInfo.LibDir)
		logger.KeyValue("Config location", platformInfo.ConfigDir)
	}

	// Create necessary directories
	logger.Progress("Creating necessary directories")
	if err := platformInfo.CreateDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Check if FixPanic Agent is already installed
	logger.Step(2, "Checking for existing installation")
	connectivityManager := connectivity.NewManager(platformInfo)
	if connectivityManager.IsFixPanicAgentInstalled() && !forceInstall {
		return fmt.Errorf("FixPanic Agent is already installed. Use --force to reinstall")
	}

	// Download FixPanic Agent binary (CORRECTED)
	logger.Step(3, "Downloading FixPanic Agent binary")
	if err := connectivityManager.DownloadFixPanicAgent("latest"); err != nil {
		return fmt.Errorf("failed to download FixPanic Agent binary: %w", err)
	}

	// Create configuration
	logger.Step(4, "Creating agent configuration")
	agentConfig := config.DefaultConfig()
	agentConfig.App.AgentID = agentID
	agentConfig.App.APIKey = agentAPIKey

	// Validate configuration
	logger.Progress("Validating configuration")
	if err := agentConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Save configuration
	configPath := platformInfo.GetConfigPath()
	if err := config.SaveConfig(agentConfig, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	logger.Success("Configuration saved to: %s", configPath)

	// Install systemd service if available
	logger.Step(5, "Setting up system service")
	if platform.IsSystemdAvailable() {
		serviceManager := service.NewManager(platformInfo)

		// Remove old service if it exists
		logger.Progress("Removing old service if it exists")
		if err := serviceManager.Uninstall(); err != nil {
			logger.Warning("Failed to remove old service: %v", err)
		}

		// Install new service
		logger.Progress("Installing systemd service")
		if err := serviceManager.Install(); err != nil {
			logger.Warning("Failed to install systemd service: %v", err)
			logger.Info("You can start the agent manually with: fixpanic agent start")
		} else {
			// Enable and start the service
			if err := serviceManager.Enable(); err != nil {
				logger.Warning("Failed to enable service: %v", err)
			}

			if err := serviceManager.Start(); err != nil {
				logger.Warning("Failed to start service: %v", err)
				logger.Info("You can start the agent manually with: fixpanic agent start")
			} else {
				logger.Success("Agent service installed and started successfully")
			}
		}
	} else {
		logger.Info("Systemd not available. You can start the agent manually with: fixpanic agent start")
	}

	logger.Separator()
	logger.Success("FixPanic agent installed successfully!")
	logger.Separator()

	logger.KeyValue("Agent ID", agentID)
	logger.KeyValue("Binary location", platformInfo.GetFixPanicAgentBinaryPath())
	logger.KeyValue("Config location", configPath)

	if platform.IsSystemdAvailable() {
		logger.Separator()
		logger.Info("The agent will start automatically on system boot.")
		logger.Info("You can manage the service with:")
		logger.Command("sudo systemctl status " + platform.GetSystemdServiceName())
		logger.Command("sudo systemctl stop " + platform.GetSystemdServiceName())
		logger.Command("sudo systemctl restart " + platform.GetSystemdServiceName())
	}

	return nil
}
