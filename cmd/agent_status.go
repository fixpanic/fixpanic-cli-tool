package cmd

import (
	"fmt"
	"os"

	"github.com/fixpanic/fixpanic-cli/internal/config"
	"github.com/fixpanic/fixpanic-cli/internal/connectivity"
	"github.com/fixpanic/fixpanic-cli/internal/logger"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/fixpanic/fixpanic-cli/internal/service"
	"github.com/spf13/cobra"
)

// agentStatusCmd represents the agent status command
var agentStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check Fixpanic agent status",
	Long: `Check the status of the Fixpanic agent on your server.
	
This command shows whether the agent is installed, running, and provides
information about the current configuration and connectivity.`,
	Example: `  # Check agent status
  fixpanic agent status`,
	RunE: runAgentStatus,
}

func init() {
	agentCmd.AddCommand(agentStatusCmd)
}

func runAgentStatus(cmd *cobra.Command, args []string) error {
	logger.Header("FixPanic Agent Status")

	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if connectivity layer is installed
	connectivityManager := connectivity.NewManager(platformInfo)
	if !connectivityManager.IsInstalled() {
		logger.Error("Agent is not installed")
		logger.Separator()
		logger.Info("To install the agent, run:")
		logger.Command("fixpanic agent install --agent-id=<your-agent-id> --api-key=<your-api-key>")
		return nil
	}

	logger.Success("Agent is installed")

	// Get FixPanic Agent version
	version, err := connectivityManager.GetFixPanicAgentVersion()
	if err != nil {
		logger.Warning("Could not determine FixPanic Agent version: %v", err)
	} else {
		logger.KeyValue("Version", version)
	}

	// Check configuration
	configPath := platformInfo.GetConfigPath()
	agentConfig, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Warning("Could not load configuration: %v", err)
	} else {
		logger.KeyValue("Configuration file", configPath)
		logger.KeyValue("Agent ID", agentConfig.App.AgentID)
		logger.KeyValue("Log level", agentConfig.Logging.Level)
	}

	// Check service status
	if platform.IsSystemdAvailable() {
		serviceManager := service.NewManager(platformInfo)

		// Check if service is enabled
		enabled, err := serviceManager.IsEnabled()
		if err != nil {
			fmt.Printf("⚠️  Could not check if service is enabled: %v\n", err)
		} else if enabled {
			fmt.Println("✅ Service is enabled for auto-start")
		} else {
			fmt.Println("⚠️  Service is not enabled for auto-start")
		}

		// Check service status
		status, err := serviceManager.Status()
		if err != nil {
			fmt.Printf("⚠️  Could not get service status: %v\n", err)
		} else {
			switch status {
			case "active":
				fmt.Println("✅ Service is running")
			case "inactive":
				fmt.Println("❌ Service is not running")
			default:
				fmt.Printf("⚠️  Service status: %s\n", status)
			}
		}
	} else {
		fmt.Println("ℹ️  Systemd not available - service management disabled")
	}

	// Check binary location
	binaryPath := platformInfo.GetFixPanicAgentBinaryPath()
	if _, err := os.Stat(binaryPath); err == nil {
		fmt.Printf("📍 Binary location: %s\n", binaryPath)
	}

	// Check log file
	logPath := fmt.Sprintf("%s/agent.log", platformInfo.LogDir)
	if _, err := os.Stat(logPath); err == nil {
		fmt.Printf("📝 Log file: %s\n", logPath)
	}

	fmt.Println("\n💡 Useful commands:")
	fmt.Println("  fixpanic agent start    - Start the agent")
	fmt.Println("  fixpanic agent stop     - Stop the agent")
	fmt.Println("  fixpanic agent logs     - View agent logs")
	fmt.Println("  fixpanic agent uninstall - Remove the agent")

	return nil
}
