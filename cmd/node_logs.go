package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
	"github.com/fixpanic/opssquad-cli-tool/internal/service"
	"github.com/spf13/cobra"
)

var logLines int
var followLogs bool

// nodeLogsCmd represents the node logs command
var nodeLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View OpsSquad node logs",
	Long: `View the logs of the OpsSquad node.
	
This command shows the node logs from systemd journal or from the log file
if systemd is not available.`,
	Example: `  # View last 50 lines of logs
  opssquad node logs
  
  # View last 100 lines of logs
  opssquad node logs --lines=100
  
  # Follow logs in real-time
  opssquad node logs --follow`,
	RunE: runNodeLogs,
}

func init() {
	nodeCmd.AddCommand(nodeLogsCmd)

	// Add flags
	nodeLogsCmd.Flags().IntVarP(&logLines, "lines", "n", 50, "Number of log lines to show")
	nodeLogsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Follow log output in real-time")
}

func runNodeLogs(cmd *cobra.Command, args []string) error {
	fmt.Println("Fetching OpsSquad node logs...")

	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Try to get logs from systemd service if available
	if platform.IsSystemdAvailable() {
		serviceManager := service.NewManager(platformInfo)

		if followLogs {
			// Follow logs in real-time
			fmt.Println("Following node logs (press Ctrl+C to stop)...")
			return followSystemdLogs(platform.GetSystemdServiceName())
		} else {
			// Get static logs
			logs, err := serviceManager.GetServiceLogs(logLines)
			if err != nil {
				fmt.Printf("Warning: could not get systemd logs: %v\n", err)
				fmt.Println("Trying to read log file directly...")
				return readLogFile(platformInfo, logLines)
			}

			if logs == "" {
				fmt.Println("No logs found for the node service.")
			} else {
				fmt.Println(logs)
			}
			return nil
		}
	}

	// Fallback: read log file directly
	fmt.Println("Systemd not available. Reading log file directly...")
	return readLogFile(platformInfo, logLines)
}

func followSystemdLogs(serviceName string) error {
	// Use journalctl to follow logs
	args := []string{"journalctl", "-u", serviceName, "-f", "--no-pager"}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to follow logs: %w", err)
	}

	return nil
}

func readLogFile(platformInfo *platform.PlatformInfo, lines int) error {
	logPath := fmt.Sprintf("%s/node.log", platformInfo.LogDir)

	// Check if log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		fmt.Printf("No log file found at: %s\n", logPath)
		fmt.Println("The node might not have been started yet, or logging might be disabled.")
		return nil
	}

	// Read the log file
	if lines > 0 {
		// Use tail to get the last N lines
		cmd := exec.Command("tail", "-n", fmt.Sprintf("%d", lines), logPath)
		output, err := cmd.Output()
		if err != nil {
			// Fallback: read entire file
			content, err := os.ReadFile(logPath)
			if err != nil {
				return fmt.Errorf("failed to read log file: %w", err)
			}
			fmt.Print(string(content))
		} else {
			fmt.Print(string(output))
		}
	} else {
		// Read entire file
		content, err := os.ReadFile(logPath)
		if err != nil {
			return fmt.Errorf("failed to read log file: %w", err)
		}
		fmt.Print(string(content))
	}

	return nil
}
