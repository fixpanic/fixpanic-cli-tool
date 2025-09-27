package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fixpanic/fixpanic-cli/internal/config"
	"github.com/fixpanic/fixpanic-cli/internal/connectivity"
	"github.com/fixpanic/fixpanic-cli/internal/logger"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/fixpanic/fixpanic-cli/internal/process"
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

// getAgentProcessInfo detects if the FixPanic Agent process is running using cross-platform process management
func getAgentProcessInfo() (running bool, pid int, err error) {
	// Create process manager for the current platform
	procManager := process.NewProcessManager()

	// Use a more targeted approach: check if the specific agent binary is running
	// We'll use the ps command approach but make it more robust
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return false, 0, fmt.Errorf("failed to execute ps command: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Look for fixpanic-connectivity-layer process (exclude grep itself and this process)
		if strings.Contains(line, "fixpanic-connectivity-layer") {
			if strings.Contains(line, "grep") || strings.Contains(line, "ps aux") {
				continue
			}

			// Extract PID from ps output
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if p, err := strconv.Atoi(fields[1]); err == nil {
					// Verify the process is actually running using our process manager
					if procManager.IsProcessRunning(p) {
						return true, p, nil
					}
				}
			}
		}
	}

	return false, 0, nil
}

// getServicePID gets the PID of the systemd service
func getServicePID() int {
	cmd := exec.Command("systemctl", "show", "-p", "MainPID", platform.GetSystemdServiceName())
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	// Parse output like "MainPID=1234"
	outputStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(outputStr, "MainPID=") {
		pidStr := strings.TrimPrefix(outputStr, "MainPID=")
		if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
			return pid
		}
	}

	return 0
}

func runAgentStatus(cmd *cobra.Command, args []string) error {
	logger.Header("FixPanic Agent Status")

	// Check if running local development version
	if rootCmd.Version == "dev" {
		fmt.Println("🚀 Running LOCAL DEVELOPMENT version (built from source)")
	}

	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if connectivity layer is installed
	connectivityManager := connectivity.NewManager(platformInfo)
	if !connectivityManager.IsFixPanicAgentInstalled() {
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

	// Check service status or process status
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
				// Try to get PID from systemctl
				if pid := getServicePID(); pid > 0 {
					fmt.Printf("🆔 Process ID: %d\n", pid)
				}
			case "inactive":
				fmt.Println("❌ Service is not running")
			default:
				fmt.Printf("⚠️  Service status: %s\n", status)
			}
		}
	} else {
		// Systemd not available, check if process is running directly using cross-platform process management
		fmt.Println("ℹ️  Systemd not available - checking process status directly")
		// Try to find the agent process by checking if any process with "fixpanic-connectivity-layer" is running
		// This is a more robust approach than the previous ps aux method
		running, pid, err := getAgentProcessInfo()
		if err != nil {
			fmt.Printf("⚠️  Could not check process status: %v\n", err)
		} else if running {
			fmt.Printf("✅ Agent is running (PID: %d)\n", pid)
		} else {
			fmt.Println("❌ Agent is not running")
		}
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
