package cmd

import (
	"github.com/spf13/cobra"
)

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage Fixpanic agents",
	Long: `Manage Fixpanic agents on your server.
	
This command group provides functionality to install, start, stop, and manage
Fixpanic agents that connect to the Fixpanic infrastructure.`,
}

func init() {
	rootCmd.AddCommand(agentCmd)
}
