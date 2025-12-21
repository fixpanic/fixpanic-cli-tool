package cmd

import (
	"github.com/spf13/cobra"
)

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage OpsSquad nodes",
	Long: `Manage OpsSquad nodes on your server.
	
This command group provides functionality to install, start, stop, and manage
OpsSquad nodes that connect to the OpsSquad infrastructure.`,
}

func init() {
	rootCmd.AddCommand(nodeCmd)
}
