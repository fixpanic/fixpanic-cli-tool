package cmd

import (
	"fmt"
	"os"

	"github.com/fixpanic/opssquad-cli-tool/internal/connectivity"
	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
	"github.com/fixpanic/opssquad-cli-tool/internal/service"
	"github.com/spf13/cobra"
)

var forceUninstall bool

// nodeUninstallCmd represents the node uninstall command
var nodeUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall OpsSquad node",
	Long: `Uninstall the OpsSquad node from your server.
	
This command removes the connectivity layer binary, configuration files,
and systemd service. Use with caution as this will completely remove
the node from your system.`,
	Example: `  # Uninstall the node
  opssquad node uninstall
  
  # Force uninstall without confirmation
  opssquad node uninstall --force`,
	RunE: runNodeUninstall,
}

func init() {
	nodeCmd.AddCommand(nodeUninstallCmd)

	// Add flags
	nodeUninstallCmd.Flags().BoolVar(&forceUninstall, "force", false, "Force uninstall without confirmation")
}

func runNodeUninstall(cmd *cobra.Command, args []string) error {
	fmt.Println("Uninstalling OpsSquad node...")

	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if OpsSquad Node is installed
	connectivityManager := connectivity.NewManager(platformInfo)
	if !connectivityManager.IsOpsSquadNodeInstalled() {
		fmt.Println("ℹ️  OpsSquad Node is not installed")
		return nil
	}

	// Confirm uninstallation unless --force is used
	if !forceUninstall {
		fmt.Println("⚠️  This will completely remove the OpsSquad node from your system.")
		fmt.Println("The following will be removed:")
		fmt.Printf("  - Binary: %s\n", platformInfo.GetBinaryPath())
		fmt.Printf("  - Configuration: %s\n", platformInfo.GetConfigPath())
		fmt.Printf("  - Service: %s\n", platform.GetSystemdServiceName())
		fmt.Printf("  - Directories: %s, %s, %s\n", platformInfo.LibDir, platformInfo.ConfigDir, platformInfo.LogDir)

		fmt.Print("\nAre you sure you want to continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Uninstallation cancelled.")
			return nil
		}
	}

	// Stop the service first
	if platform.IsSystemdAvailable() {
		serviceManager := service.NewManager(platformInfo)

		// Check if service is running
		status, err := serviceManager.Status()
		if err == nil && status == "active" {
			fmt.Println("Stopping node service...")
			if err := serviceManager.Stop(); err != nil {
				fmt.Printf("Warning: failed to stop service: %v\n", err)
			}
		}

		// Uninstall service
		fmt.Println("Removing systemd service...")
		if err := serviceManager.Uninstall(); err != nil {
			fmt.Printf("Warning: failed to uninstall service: %v\n", err)
		}
	}

	// Remove OpsSquad Node binary
	fmt.Println("Removing OpsSquad Node binary...")
	if err := connectivityManager.RemoveOpsSquadNode(); err != nil {
		fmt.Printf("Warning: failed to remove binary: %v\n", err)
	}

	// Remove configuration file
	configPath := platformInfo.GetConfigPath()
	fmt.Printf("Removing configuration file: %s\n", configPath)
	if err := os.Remove(configPath); err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to remove configuration file: %v\n", err)
		}
	}

	// Remove directories (only if empty)
	dirs := []string{
		platformInfo.LibDir,
		platformInfo.ConfigDir,
		platformInfo.LogDir,
	}

	for _, dir := range dirs {
		if err := os.Remove(dir); err != nil {
			// Directory not empty or doesn't exist, which is fine
			continue
		}
		fmt.Printf("Removed empty directory: %s\n", dir)
	}

	fmt.Println("\n✅ OpsSquad node uninstalled successfully!")
	fmt.Println("The node has been completely removed from your system.")

	return nil
}
