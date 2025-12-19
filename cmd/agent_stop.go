package cmd

import (
	"fmt"

	"github.com/fixpanic/fixpanic-cli/internal/logger"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/fixpanic/fixpanic-cli/internal/process"
	"github.com/fixpanic/fixpanic-cli/internal/service"
	"github.com/spf13/cobra"
)

var agentStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the FixPanic Agent",
	Long:  `Stop the FixPanic Agent service that is running in the background.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Header("Stopping FixPanic Agent")

		platformInfo, err := platform.GetPlatformInfo()
		if err != nil {
			return fmt.Errorf("failed to get platform info: %w", err)
		}

		stoppedSomething := false

		// 1. Try systemd stop if available and usable
		if platform.IsSystemdAvailable() {
			serviceManager := service.NewManager(platformInfo)
			if serviceManager.IsUsable() {
				status, _ := serviceManager.Status()
				if status == "active" || status == "activating" {
					logger.Step(1, "Stopping systemd service")
					if err := serviceManager.Stop(); err != nil {
						fmt.Printf("Warning: failed to stop service: %v\n", err)
					} else {
						fmt.Println("✅ Service stopped successfully")
						stoppedSomething = true
					}
				}
			}
		}

		// 2. Always check for direct processes (cleanup/fallback)
		logger.Step(2, "Checking for background processes")
		pids, err := getAllAgentProcessPIDs()
		if err != nil {
			return fmt.Errorf("failed to check agent processes: %w", err)
		}

		if len(pids) > 0 {
			fmt.Printf("Found %d running process(es)\n", len(pids))
			procManager := process.NewProcessManager()
			stoppedCount := 0
			for _, pid := range pids {
				fmt.Printf("Stopping PID: %d... ", pid)
				if err := procManager.StopProcess(pid); err != nil {
					fmt.Printf("Failed: %v\n", err)
				} else {
					fmt.Println("Done")
					stoppedCount++
				}
			}
			if stoppedCount > 0 {
				stoppedSomething = true
			}
		} else {
			fmt.Println("No background processes found")
		}

		if !stoppedSomething {
			fmt.Println("\n⚠️  FixPanic Agent is not running (checked service and background processes)")
		} else {
			logger.Success("Agent stopped")
		}

		return nil
	},
}

func init() {
	agentCmd.AddCommand(agentStopCmd)
}
