package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fixpanic/opssquad-cli-tool/internal/logger"
	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
	"github.com/spf13/cobra"
)

// nodeUnlockCmd represents the node unlock command
var nodeUnlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Remove emergency lock: restore session-based access control",
	Long: `Remove the local lock file to restore normal session-based operation.

After unlocking, the node will follow the organization session state:
commands are accepted only when an active session exists.`,
	Example: `  # Unlock the node
  opssquad node unlock`,
	RunE: runNodeUnlock,
}

func init() {
	nodeCmd.AddCommand(nodeUnlockCmd)
}

func runNodeUnlock(cmd *cobra.Command, args []string) error {
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	configDir := filepath.Dir(platformInfo.GetConfigPath())
	lockPath := filepath.Join(configDir, lockFileName)

	// Check if lock exists
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		logger.Info("Node is not locked. No action needed.")
		return nil
	}

	// Remove lock file
	if err := os.Remove(lockPath); err != nil {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	logger.Success("Local lock removed")
	fmt.Println("Node will now follow organization session state.")

	return nil
}
