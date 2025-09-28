package cmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fixpanic/fixpanic-cli/internal/connectivity"
	"github.com/fixpanic/fixpanic-cli/internal/logger"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/fixpanic/fixpanic-cli/internal/process"
	"github.com/fixpanic/fixpanic-cli/internal/service"
	"github.com/spf13/cobra"
)

// agentStartCmd represents the agent start command
var agentStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Fixpanic agent",
	Long: `Start the Fixpanic agent service.
	
This command starts the agent service using systemd if available, or runs the
connectivity layer binary directly if systemd is not available.`,
	Example: `  # Start the agent
  fixpanic agent start`,
	RunE: runAgentStart,
}

func init() {
	agentCmd.AddCommand(agentStartCmd)
}

func runAgentStart(cmd *cobra.Command, args []string) error {
	logger.Header("Starting FixPanic Agent")

	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Validate agent installation
	connectivityManager, err := validateAgentInstall(platformInfo)
	if err != nil {
		return err
	}

	// Clean up old agents
	logger.Step(2, "Checking for existing agent processes")
	if err := cleanUpOldAgents(); err != nil {
		return err
	}

	// Start the agent service
	return startAgentService(platformInfo, connectivityManager)
}

// validateAgentInstall checks if the agent is installed
func validateAgentInstall(platformInfo *platform.PlatformInfo) (*connectivity.Manager, error) {
	logger.Step(1, "Checking agent installation")
	connectivityManager := connectivity.NewManager(platformInfo)

	if !connectivityManager.IsFixPanicAgentInstalled() {
		return nil, fmt.Errorf("FixPanic Agent not installed. Run 'fixpanic agent install' first")
	}

	logger.Success("Agent installation verified")
	return connectivityManager, nil
}

// cleanUpOldAgents stops any existing agent processes before starting a new one
func cleanUpOldAgents() error {
	existingPIDs, err := getAllAgentProcessPIDs()
	if err != nil {
		return fmt.Errorf("failed to check for existing agent processes: %w", err)
	}

	if len(existingPIDs) > 0 {
		fmt.Printf("⚠️  Found %d existing agent process(es) running:\n", len(existingPIDs))
		for _, pid := range existingPIDs {
			fmt.Printf("   - PID: %d\n", pid)
		}
		fmt.Println("🛑 Stopping existing processes before starting new agent...")

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
			return fmt.Errorf("failed to stop any existing agent processes")
		}

		fmt.Printf("✅ Stopped %d existing process(es)\n", stoppedCount)
		fmt.Println() // Empty line for better readability
	}

	return nil
}

// startAgentService starts the agent using systemd if available, or directly if not
func startAgentService(platformInfo *platform.PlatformInfo, connectivityManager *connectivity.Manager) error {
	binaryPath := platformInfo.GetFixPanicAgentBinaryPath()

	// Try to use systemd service if available
	logger.Step(3, "Starting agent service")
	if platform.IsSystemdAvailable() {
		serviceManager := service.NewManager(platformInfo)

		// Check current status
		logger.Progress("Checking service status")
		status, err := serviceManager.Status()
		if err != nil {
			fmt.Printf("Warning: could not check service status: %v\n", err)
		} else if status == "active" {
			fmt.Println("✅ Agent service is already running")
			return nil
		}

		// Start the service
		if err := serviceManager.Start(); err != nil {
			return fmt.Errorf("failed to start service: %w", err)
		}

		fmt.Println("✅ Agent service started successfully")
		fmt.Printf("Service: %s\n", platform.GetSystemdServiceName())

		// Show how to check status
		fmt.Println("\nYou can check the status with:")
		fmt.Printf("  sudo systemctl status %s\n", platform.GetSystemdServiceName())

		return nil
	}

	// Use cross-platform process manager for direct process execution
	configPath := platformInfo.GetConfigPath()

	fmt.Printf("Starting: %s --config %s\n", binaryPath, configPath)

	// Create process manager for the current platform
	procManager := process.NewProcessManager()

	// Start the agent process
	procInfo, err := procManager.StartProcess(process.ProcessConfig{
		BinaryPath: binaryPath,
		Args:       []string{"--config", configPath},
		Detach:     true,
	})
	if err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	fmt.Println("✅ Agent started successfully in background")
	fmt.Printf("Process PID: %d\n", procInfo.PID)

	return nil
}

// getAllAgentProcessPIDs returns all PIDs of running FixPanic Agent processes
func getAllAgentProcessPIDs() ([]int, error) {
	var pids []int

	// Create process manager for the current platform
	procManager := process.NewProcessManager()

	// Use ps command to find all fixpanic-connectivity-layer processes
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute ps command: %w", err)
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
