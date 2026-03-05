package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fixpanic/opssquad-cli-tool/internal/logger"
	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
	"github.com/spf13/cobra"
)

const lockFileName = ".opssquad-locked"

// nodeLockCmd represents the node lock command
var nodeLockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Emergency lock: block all command execution on this node",
	Long: `Create a local lock file that blocks all command execution regardless of session state.

This is an offline emergency kill switch. When the lock file exists, the node will
reject all execution requests even if an organization session is active.

Use 'opssquad node unlock' to remove the lock and return to normal session-based operation.`,
	Example: `  # Lock the node immediately
  opssquad node lock`,
	RunE: runNodeLock,
}

func init() {
	nodeCmd.AddCommand(nodeLockCmd)
}

func runNodeLock(cmd *cobra.Command, args []string) error {
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	configDir := filepath.Dir(platformInfo.GetConfigPath())
	lockPath := filepath.Join(configDir, lockFileName)

	// Check if already locked
	if _, err := os.Stat(lockPath); err == nil {
		logger.Warning("Node is already locked")
		logger.KeyValue("Lock file", lockPath)
		return nil
	}

	// Create lock file
	if err := os.WriteFile(lockPath, []byte("locked\n"), 0644); err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	logger.Success("Node locked")
	fmt.Println("All command execution blocked regardless of session state.")
	logger.KeyValue("Lock file", lockPath)
	fmt.Println("\nTo unlock: opssquad node unlock")

	return nil
}
