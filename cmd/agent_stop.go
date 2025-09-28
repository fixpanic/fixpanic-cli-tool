package cmd

import (
	"fmt"

	"github.com/fixpanic/fixpanic-cli/internal/process"
	"github.com/spf13/cobra"
)

var agentStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the FixPanic Agent",
	Long:  `Stop the FixPanic Agent service that is running in the background.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get all running agent processes
		pids, err := getAllAgentProcessPIDs()
		if err != nil {
			return fmt.Errorf("failed to check agent status: %w", err)
		}

		if len(pids) == 0 {
			fmt.Println("FixPanic Agent is not running")
			return nil
		}

		// Create process manager for the current platform
		procManager := process.NewProcessManager()

		// Stop all agent processes
		stoppedCount := 0
		for _, pid := range pids {
			fmt.Printf("Stopping FixPanic Agent (PID: %d)...\n", pid)
			if err := procManager.StopProcess(pid); err != nil {
				fmt.Printf("Warning: failed to stop process %d: %v\n", pid, err)
			} else {
				stoppedCount++
			}
		}

		if stoppedCount == 0 {
			return fmt.Errorf("failed to stop any agent processes")
		}

		if stoppedCount == 1 {
			fmt.Println("FixPanic Agent stopped successfully")
		} else {
			fmt.Printf("FixPanic Agent stopped successfully (%d processes stopped)\n", stoppedCount)
		}
		return nil
	},
}

func init() {
	agentCmd.AddCommand(agentStopCmd)
}
