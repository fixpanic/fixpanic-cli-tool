package cmd

import (
	"fmt"

	"github.com/fixpanic/opssquad-cli-tool/internal/config"
	"github.com/fixpanic/opssquad-cli-tool/internal/connectivity"
	"github.com/fixpanic/opssquad-cli-tool/internal/logger"
	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
	"github.com/fixpanic/opssquad-cli-tool/internal/service"
	"github.com/spf13/cobra"
)

var (
	nodeID       string
	nodeToken    string
	forceInstall bool
)

// nodeInstallCmd represents the node install command
var nodeInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install OpsSquad node",
	Long: `Install the OpsSquad node on your server.

This command downloads and installs the connectivity layer binary, creates the
necessary configuration files, and sets up the systemd service for automatic
startup.`,
	Example: `  # Install with node credentials
	 opssquad node install --node-id="node_123" --token="fp_abc123xyz"

	 # Force reinstall
	 opssquad node install --node-id="node_123" --token="fp_abc123xyz" --force`,
	RunE: runNodeInstall,
}

func init() {
	nodeCmd.AddCommand(nodeInstallCmd)

	// Add flags
	nodeInstallCmd.Flags().StringVar(&nodeID, "node-id", "", "Node ID from OpsSquad dashboard (required)")
	nodeInstallCmd.Flags().StringVar(&nodeToken, "token", "", "Node token from OpsSquad dashboard (required)")
	nodeInstallCmd.Flags().BoolVar(&forceInstall, "force", false, "Force reinstall even if node is already installed")

	// Mark required flags
	_ = nodeInstallCmd.MarkFlagRequired("node-id")
	_ = nodeInstallCmd.MarkFlagRequired("token")
}

func runNodeInstall(cmd *cobra.Command, args []string) error {
	logger.Header("Installing OpsSquad Node")

	// Get platform information
	logger.Step(1, "Detecting platform and configuration")
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if running as root for system-wide installation
	if !platformInfo.IsRoot {
		logger.Warning("Running as non-root user. Node will be installed in user directories.")
		logger.KeyValue("Binary location", platformInfo.LibDir)
		logger.KeyValue("Config location", platformInfo.ConfigDir)
	}

	// Create necessary directories
	logger.Progress("Creating necessary directories")
	if err := platformInfo.CreateDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Check if OpsSquad Node is already installed
	logger.Step(2, "Checking for existing installation")
	connectivityManager := connectivity.NewManager(platformInfo)
	if connectivityManager.IsBinaryInstalled() && !forceInstall {
		return fmt.Errorf("OpsSquad Node is already installed. Use --force to reinstall")
	}

	// Ensure latest node binary (auto-update)
	logger.Step(3, "Ensuring latest node binary")
	if err := connectivityManager.EnsureLatestNode(); err != nil {
		return fmt.Errorf("failed to ensure latest node binary: %w", err)
	}

	// Create configuration
	logger.Step(4, "Creating node configuration")
	// default config with TLS enabled
	nodeConfig := config.DefaultConfig(config.DefaultConfigOptions{
		TLSEnabled: true, // TLS Enable by default for security
		NodeID:     nodeID,
		Token:      nodeToken,
	})

	// Validate configuration
	logger.Progress("Validating configuration")
	if err := nodeConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Save configuration
	configPath := platformInfo.GetConfigPath()
	if err := config.SaveConfig(nodeConfig, configPath); err != nil {
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
			logger.Info("You can start the node manually with: opssquad node start")
		} else {
			// Enable and start the service
			if err := serviceManager.Enable(); err != nil {
				logger.Warning("Failed to enable service: %v", err)
			}

			if err := serviceManager.Start(); err != nil {
				logger.Warning("Failed to start service: %v", err)
				logger.Info("You can start the node manually with: opssquad node start")
			} else {
				logger.Success("Node service installed and started successfully")
			}
		}
	} else {
		logger.Info("Systemd not available. You can start the node manually with: opssquad node start")
	}

	logger.Separator()
	logger.Success("OpsSquad node installed successfully!")
	logger.Separator()

	logger.KeyValue("Node ID", nodeID)
	logger.KeyValue("Binary location", platformInfo.GetBinaryPath())
	logger.KeyValue("Config location", configPath)

	if platform.IsSystemdAvailable() {
		logger.Separator()
		logger.Info("The node will start automatically on system boot.")
		logger.Info("You can manage the service with:")
		logger.Command("sudo systemctl status " + platform.GetSystemdServiceName())
		logger.Command("sudo systemctl stop " + platform.GetSystemdServiceName())
		logger.Command("sudo systemctl restart " + platform.GetSystemdServiceName())
	}

	return nil
}
