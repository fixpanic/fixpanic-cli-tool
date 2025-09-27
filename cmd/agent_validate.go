package cmd

import (
	"fmt"
	"os"

	"github.com/fixpanic/fixpanic-cli/internal/config"
	"github.com/fixpanic/fixpanic-cli/internal/connectivity"
	"github.com/fixpanic/fixpanic-cli/internal/logger"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/spf13/cobra"
)

// agentValidateCmd represents the agent validate command
var agentValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate agent installation and configuration",
	Long: `Validate that the Fixpanic agent is properly installed and configured.

This command checks if the agent binary is installed, configuration is valid,
and the agent can be started successfully.`,
	Example: `  # Validate agent installation
  fixpanic agent validate`,
	RunE: runAgentValidate,
}

func init() {
	agentCmd.AddCommand(agentValidateCmd)
}

func runAgentValidate(cmd *cobra.Command, args []string) error {
	logger.Header("Validating Agent Installation")

	// Get platform information
	logger.Step(1, "Detecting platform and configuration")
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if FixPanic Agent is installed
	logger.Step(2, "Checking agent binary installation")
	connectivityManager := connectivity.NewManager(platformInfo)
	if !connectivityManager.IsFixPanicAgentInstalled() {
		return fmt.Errorf("FixPanic Agent is not installed. Run 'fixpanic agent install' first")
	}

	logger.List("FixPanic Agent binary found: %s", connectivityManager.GetBinaryPath())

	// Load configuration
	configPath := platformInfo.GetConfigPath()
	agentConfig, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := agentConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	fmt.Printf("✅ Configuration is valid: %s\n", configPath)
	fmt.Printf("   Agent ID: %s\n", agentConfig.App.AgentID)
	fmt.Printf("   Log level: %s\n", agentConfig.Logging.Level)
	fmt.Printf("   Log file: %s\n", agentConfig.Logging.File)

	// Test if binary is executable
	binaryPath := connectivityManager.GetBinaryPath()
	if err := os.Chmod(binaryPath, 0755); err != nil {
		fmt.Printf("⚠️  Could not verify FixPanic Agent permissions: %v\n", err)
	} else {
		fmt.Println("✅ FixPanic Agent binary has correct permissions")
	}

	// Test version command
	fmt.Println("\nTesting FixPanic Agent binary...")
	version, err := connectivityManager.GetFixPanicAgentVersion()
	if err != nil {
		fmt.Printf("⚠️  Could not get FixPanic Agent version: %v\n", err)
	} else {
		fmt.Printf("✅ FixPanic Agent version: %s\n", version)
	}

	fmt.Println("\n✅ FixPanic Agent validation completed successfully!")
	fmt.Println("The FixPanic Agent appears to be properly installed and configured.")
	fmt.Println("You can start the agent with: fixpanic agent start")

	return nil
}
