package cmd

import (
	"fmt"
	"os"

	"github.com/fixpanic/fixpanic-cli/internal/connectivity"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/fixpanic/fixpanic-cli/internal/service"
	"github.com/spf13/cobra"
)

var forceUninstall bool

// agentUninstallCmd represents the agent uninstall command
var agentUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Fixpanic agent",
	Long: `Uninstall the Fixpanic agent from your server.
	
This command removes the connectivity layer binary, configuration files,
and systemd service. Use with caution as this will completely remove
the agent from your system.`,
	Example: `  # Uninstall the agent
  fixpanic agent uninstall
  
  # Force uninstall without confirmation
  fixpanic agent uninstall --force`,
	RunE: runAgentUninstall,
}

func init() {
	agentCmd.AddCommand(agentUninstallCmd)

	// Add flags
	agentUninstallCmd.Flags().BoolVar(&forceUninstall, "force", false, "Force uninstall without confirmation")
}

func runAgentUninstall(cmd *cobra.Command, args []string) error {
	fmt.Println("Uninstalling Fixpanic agent...")

	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if FixPanic Agent is installed
	connectivityManager := connectivity.NewManager(platformInfo)
	if !connectivityManager.IsFixPanicAgentInstalled() {
		fmt.Println("ℹ️  FixPanic Agent is not installed")
		return nil
	}

	// Confirm uninstallation unless --force is used
	if !forceUninstall {
		fmt.Println("⚠️  This will completely remove the Fixpanic agent from your system.")
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
			fmt.Println("Stopping agent service...")
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

	// Remove FixPanic Agent binary
	fmt.Println("Removing FixPanic Agent binary...")
	if err := connectivityManager.RemoveFixPanicAgent(); err != nil {
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

	fmt.Println("\n✅ Fixpanic agent uninstalled successfully!")
	fmt.Println("The agent has been completely removed from your system.")

	return nil
}
