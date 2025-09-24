package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/fixpanic/fixpanic-cli/internal/service"
	"github.com/spf13/cobra"
)

var logLines int
var followLogs bool

// agentLogsCmd represents the agent logs command
var agentLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View Fixpanic agent logs",
	Long: `View the logs of the Fixpanic agent.
	
This command shows the agent logs from systemd journal or from the log file
if systemd is not available.`,
	Example: `  # View last 50 lines of logs
  fixpanic agent logs
  
  # View last 100 lines of logs
  fixpanic agent logs --lines=100
  
  # Follow logs in real-time
  fixpanic agent logs --follow`,
	RunE: runAgentLogs,
}

func init() {
	agentCmd.AddCommand(agentLogsCmd)

	// Add flags
	agentLogsCmd.Flags().IntVarP(&logLines, "lines", "n", 50, "Number of log lines to show")
	agentLogsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Follow log output in real-time")
}

func runAgentLogs(cmd *cobra.Command, args []string) error {
	fmt.Println("Fetching Fixpanic agent logs...")

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
			fmt.Println("Following agent logs (press Ctrl+C to stop)...")
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
				fmt.Println("No logs found for the agent service.")
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
	logPath := fmt.Sprintf("%s/agent.log", platformInfo.LogDir)

	// Check if log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		fmt.Printf("No log file found at: %s\n", logPath)
		fmt.Println("The agent might not have been started yet, or logging might be disabled.")
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
