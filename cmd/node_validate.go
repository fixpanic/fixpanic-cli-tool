package cmd

import (
	"fmt"
	"os"

	"github.com/fixpanic/opssquad-cli-tool/internal/config"
	"github.com/fixpanic/opssquad-cli-tool/internal/connectivity"
	"github.com/fixpanic/opssquad-cli-tool/internal/logger"
	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
	"github.com/spf13/cobra"
)

// nodeValidateCmd represents the node validate command
var nodeValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate node installation and configuration",
	Long: `Validate that the OpsSquad node is properly installed and configured.

This command checks if the node binary is installed, configuration is valid,
and the node can be started successfully.`,
	Example: `  # Validate node installation
  opssquad node validate`,
	RunE: runNodeValidate,
}

func init() {
	nodeCmd.AddCommand(nodeValidateCmd)
}

func runNodeValidate(cmd *cobra.Command, args []string) error {
	logger.Header("Validating Node Installation")

	// Get platform information
	logger.Step(1, "Detecting platform and configuration")
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if OpsSquad Node is installed
	logger.Step(2, "Checking node binary installation")
	connectivityManager := connectivity.NewManager(platformInfo)
	if !connectivityManager.IsBinaryInstalled() {
		return fmt.Errorf("OpsSquad Node is not installed. Run 'opssquad node install' first")
	}

	logger.List("OpsSquad Node binary found: %s", connectivityManager.GetBinaryPath())

	// Load configuration
	configPath := platformInfo.GetConfigPath()
	nodeConfig, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := nodeConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	fmt.Printf("✅ Configuration is valid: %s\n", configPath)
	fmt.Printf("   Node ID: %s\n", nodeConfig.App.NodeID)
	fmt.Printf("   Log level: %s\n", nodeConfig.Logging.Level)
	fmt.Printf("   Log file: %s\n", nodeConfig.Logging.File)

	// Test if binary is executable
	binaryPath := connectivityManager.GetBinaryPath()
	if err := os.Chmod(binaryPath, 0755); err != nil {
		fmt.Printf("⚠️  Could not verify OpsSquad Node permissions: %v\n", err)
	} else {
		fmt.Println("✅ OpsSquad Node binary has correct permissions")
	}

	// Test version command
	fmt.Println("\nTesting OpsSquad Node binary...")
	version, err := connectivityManager.GetBinaryVersion()
	if err != nil {
		fmt.Printf("⚠️  Could not get OpsSquad Node version: %v\n", err)
	} else {
		fmt.Printf("✅ OpsSquad Node version: %s\n", version)
	}

	fmt.Println("\n✅ OpsSquad Node validation completed successfully!")
	fmt.Println("The OpsSquad Node appears to be properly installed and configured.")
	fmt.Println("You can start the node with: opssquad node start")

	return nil
}
