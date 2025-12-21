package cmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fixpanic/opssquad-cli-tool/internal/connectivity"
	"github.com/fixpanic/opssquad-cli-tool/internal/logger"
	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
	"github.com/fixpanic/opssquad-cli-tool/internal/process"
	"github.com/fixpanic/opssquad-cli-tool/internal/service"
	"github.com/spf13/cobra"
)

// nodeStartCmd represents the node start command
var nodeStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start OpsSquad node",
	Long: `Start the OpsSquad node service.
	
This command starts the node service using systemd if available, or runs the
connectivity layer binary directly if systemd is not available.`,
	Example: `  # Start the node
  opssquad node start`,
	RunE: runNodeStart,
}

func init() {
	nodeCmd.AddCommand(nodeStartCmd)
}

func runNodeStart(cmd *cobra.Command, args []string) error {
	logger.Header("Starting OpsSquad Node")

	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check environment and attempt auto-configuration
	if warnings := platformInfo.AutoConfigureEnvironment(); len(warnings) > 0 {
		for _, warning := range warnings {
			fmt.Printf("⚠️  Warning: %s\n", warning)
		}
	}

	// Validate node installation
	connectivityManager, err := validateNodeInstall(platformInfo)
	if err != nil {
		return err
	}

	// Clean up old nodes
	logger.Step(2, "Checking for existing node processes")
	if err := cleanUpOldNodes(); err != nil {
		return err
	}

	// Start the node service
	return startNodeService(platformInfo, connectivityManager)
}

// validateNodeInstall checks if the node is installed
func validateNodeInstall(platformInfo *platform.PlatformInfo) (*connectivity.Manager, error) {
	logger.Step(1, "Checking node installation")
	connectivityManager := connectivity.NewManager(platformInfo)

	if !connectivityManager.IsOpsSquadNodeInstalled() {
		return nil, fmt.Errorf("OpsSquad Node not installed. Run 'opssquad node install' first")
	}

	logger.Success("Node installation verified")
	return connectivityManager, nil
}

// cleanUpOldNodes stops any existing node processes before starting a new one
func cleanUpOldNodes() error {
	existingPIDs, err := getAllNodeProcessPIDs()
	if err != nil {
		return fmt.Errorf("failed to check for existing node processes: %w", err)
	}

	if len(existingPIDs) > 0 {
		fmt.Printf("⚠️  Found %d existing node process(es) running:\n", len(existingPIDs))
		for _, pid := range existingPIDs {
			fmt.Printf("   - PID: %d\n", pid)
		}
		fmt.Println("🛑 Stopping existing processes before starting new node...")

		// Stop all existing processes
		procManager := process.NewProcessManager()
		stoppedCount := 0
		for _, pid := range existingPIDs {
			if err := procManager.StopProcess(pid); err != nil {
				fmt.Printf("⚠️  Warning: failed to stop process %d: %v\n", pid, err)
			} else {
				stoppedCount++
			}
		}

		if stoppedCount == 0 {
			return fmt.Errorf("failed to stop any existing node processes")
		}

		fmt.Printf("✅ Stopped %d existing process(es)\n", stoppedCount)
		fmt.Println() // Empty line for better readability
	}

	return nil
}

// startNodeService starts the node using systemd if available, or directly if not
func startNodeService(platformInfo *platform.PlatformInfo, connectivityManager *connectivity.Manager) error {
	binaryPath := platformInfo.GetOpsSquadNodeBinaryPath()
	useSystemd := false

	// Try to use systemd service if available AND usable
	if platform.IsSystemdAvailable() {
		serviceManager := service.NewManager(platformInfo)
		if serviceManager.IsUsable() {
			useSystemd = true
		} else {
			logger.Warning("Systemd is available but not usable (missing privileges or socket). Falling back to direct process execution.")
		}
	}

	if useSystemd {
		logger.Step(3, "Starting node service")
		serviceManager := service.NewManager(platformInfo)

		// Check current status
		logger.Progress("Checking service status")
		status, err := serviceManager.Status()
		if err != nil {
			fmt.Printf("Warning: could not check service status: %v\n", err)
		} else if status == "active" {
			fmt.Println("✅ Node service is already running")
			return nil
		}

		// Start the service
		if err := serviceManager.Start(); err != nil {
			return fmt.Errorf("failed to start service: %w", err)
		}

		fmt.Println("✅ Node service started successfully")
		fmt.Printf("Service: %s\n", platform.GetSystemdServiceName())

		// Show how to check status
		fmt.Println("\nYou can check the status with:")
		fmt.Printf("  sudo systemctl status %s\n", platform.GetSystemdServiceName())

		return nil
	}

	// Use cross-platform process manager for direct process execution
	logger.Step(3, "Starting node process (background)")
	configPath := platformInfo.GetConfigPath()

	fmt.Printf("Starting: %s --config %s\n", binaryPath, configPath)

	// Create process manager for the current platform
	procManager := process.NewProcessManager()

	// Start the node process
	procInfo, err := procManager.StartProcess(process.ProcessConfig{
		BinaryPath: binaryPath,
		Args:       []string{"--config", configPath},
		Detach:     true,
	})
	if err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	fmt.Println("✅ Node started successfully in background")
	fmt.Printf("Process PID: %d\n", procInfo.PID)

	return nil
}

// getAllNodeProcessPIDs returns all PIDs of running OpsSquad Node processes
func getAllNodeProcessPIDs() ([]int, error) {
	var pids []int

	// Create process manager for the current platform
	procManager := process.NewProcessManager()

	// Use ps command to find all opssquad-connectivity-layer processes
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute ps command: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Look for opssquad-connectivity-layer process (exclude grep itself and this process)
		if strings.Contains(line, "opssquad-connectivity-layer") {
			if strings.Contains(line, "grep") || strings.Contains(line, "ps aux") {
				continue
			}

			// Extract PID from ps output
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if pid, err := strconv.Atoi(fields[1]); err == nil {
					// Verify the process is actually running using our process manager
					if procManager.IsProcessRunning(pid) {
						pids = append(pids, pid)
					}
				}
			}
		}
	}

	return pids, nil
}
