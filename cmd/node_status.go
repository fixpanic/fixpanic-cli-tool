package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fixpanic/opssquad-cli-tool/internal/config"
	"github.com/fixpanic/opssquad-cli-tool/internal/connectivity"
	"github.com/fixpanic/opssquad-cli-tool/internal/logger"
	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
	"github.com/fixpanic/opssquad-cli-tool/internal/process"
	"github.com/fixpanic/opssquad-cli-tool/internal/service"
	"github.com/spf13/cobra"
)

// nodeStatusCmd represents the node status command
var nodeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check OpsSquad node status",
	Long: `Check the status of the OpsSquad node on your server.
	
This command shows whether the node is installed, running, and provides
information about the current configuration and connectivity.`,
	Example: `  # Check node status
  opssquad node status`,
	RunE: runNodeStatus,
}

func init() {
	nodeCmd.AddCommand(nodeStatusCmd)
}

// getNodeProcessInfo detects if the OpsSquad Node process is running using cross-platform process management
func getNodeProcessInfo() (running bool, pid int, err error) {
	// Create process manager for the current platform
	procManager := process.NewProcessManager()

	// Use a more targeted approach: check if the specific node binary is running
	// We'll use the ps command approach but make it more robust
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return false, 0, fmt.Errorf("failed to execute ps command: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Look for opssquad-connectivity-layer process (exclude grep itself and this process)
		if strings.Contains(line, "opssquad-connectivity-layer") {
			if strings.Contains(line, "grep") || strings.Contains(line, "ps aux") {
				continue
			}

			// Extract PID from ps output
			// Handle both GNU ps (USER PID ...) and busybox ps (PID USER ...) formats
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				var pid int
				var err error

				// Try to parse first field as PID (busybox format: PID USER TIME COMMAND)
				pid, err = strconv.Atoi(fields[0])
				if err != nil {
					// If first field isn't a number, try second field (GNU format: USER PID %CPU ...)
					pid, err = strconv.Atoi(fields[1])
					if err != nil {
						continue
					}
				}

				// Verify the process is actually running using our process manager
				if procManager.IsProcessRunning(pid) {
					return true, pid, nil
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

func runNodeStatus(cmd *cobra.Command, args []string) error {
	logger.Header("OpsSquad Node Status")

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
	if !connectivityManager.IsBinaryInstalled() {
		logger.Error("Node is not installed")
		logger.Separator()
		logger.Info("To install the node, run:")
		logger.Command("opssquad node install --node-id=<your-node-id> --token=<your-token>")
		return nil
	}

	logger.Success("Node is installed")

	// Get OpsSquad Node version
	version, err := connectivityManager.GetBinaryVersion()
	if err != nil {
		logger.Warning("Could not determine OpsSquad Node version: %v", err)
	} else {
		logger.KeyValue("Version", version)
	}

	// Check configuration
	configPath := platformInfo.GetConfigPath()
	logger.KeyValue("Configuration file", configPath)

	nodeConfig, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Warning("Could not load configuration: %v", err)
	} else {
		logger.KeyValue("Node ID", nodeConfig.App.NodeID)
		logger.KeyValue("Log level", nodeConfig.Logging.Level)
	}

	// Check service status or process status
	var isRunning bool
	var serviceStatus string
	var mode string = "None"
	var pid int

	// 1. Check Systemd (if available)
	if platform.IsSystemdAvailable() {
		serviceManager := service.NewManager(platformInfo)

		// Check enabled status
		if serviceManager.IsUsable() {
			enabled, err := serviceManager.IsEnabled()
			if err == nil && enabled {
				fmt.Println("✅ Service is enabled for auto-start")
			} else {
				fmt.Println("ℹ️  Service is not enabled for auto-start")
			}

			// Check status
			s, err := serviceManager.Status()
			if err == nil {
				serviceStatus = s
				if s == "active" {
					isRunning = true
					mode = "Systemd Service"
					// Try to get PID
					pid = getServicePID()
				}
			}
		}
	}

	// 2. If not running via Systemd, check direct process (Fallback/Legacy)
	if !isRunning {
		// Systemd not active or not available, check direct process
		var err error
		isRunning, pid, err = getNodeProcessInfo()
		if err != nil {
			fmt.Printf("⚠️  Could not check process status: %v\n", err)
		} else if isRunning {
			mode = "Background Process"
		}
	}

	// Report Status
	if isRunning {
		logger.Success("Node is RUNNING")
		logger.KeyValue("Mode", mode)
		if pid > 0 {
			logger.KeyValue("PID", fmt.Sprintf("%d", pid))
		}
		if serviceStatus != "" && mode != "Systemd Service" {
			// Inform user that systemd thinks it's stopped, but process is running
			fmt.Printf("ℹ️  Note: Systemd service status is '%s', but node is running as a background process.\n", serviceStatus)
		}
	} else {
		// Not running
		fmt.Println("❌ Node is STOPPED")
		if serviceStatus != "" {
			fmt.Printf("   Service status: %s\n", serviceStatus)
		}
	}

	// Check binary location
	binaryPath := platformInfo.GetBinaryPath()
	if _, err := os.Stat(binaryPath); err == nil {
		fmt.Printf("📍 Binary location: %s\n", binaryPath)
	}

	// Check log file
	logPath := fmt.Sprintf("%s/node.log", platformInfo.LogDir)
	if _, err := os.Stat(logPath); err == nil {
		fmt.Printf("📝 Log file: %s\n", logPath)
	}

	// Check local lock status
	configDir := filepath.Dir(platformInfo.GetConfigPath())
	lockPath := filepath.Join(configDir, lockFileName)
	if _, err := os.Stat(lockPath); err == nil {
		fmt.Println("🔒 Local lock: ACTIVE (all execution blocked)")
	} else {
		fmt.Println("🔓 Local lock: none")
	}

	fmt.Println("\n💡 Useful commands:")
	fmt.Println("  opssquad node start    - Start the node")
	fmt.Println("  opssquad node stop     - Stop the node")
	fmt.Println("  opssquad node logs     - View node logs")
	fmt.Println("  opssquad node lock     - Emergency lock (block all execution)")
	fmt.Println("  opssquad node unlock   - Remove emergency lock")
	fmt.Println("  opssquad node uninstall - Remove the node")

	return nil
}
