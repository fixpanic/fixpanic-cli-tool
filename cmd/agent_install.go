package cmd

import (
	"fmt"

	"github.com/fixpanic/fixpanic-cli/internal/config"
	"github.com/fixpanic/fixpanic-cli/internal/connectivity"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/fixpanic/fixpanic-cli/internal/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	agentID      string
	agentAPIKey  string
	socketServer string
	forceInstall bool
)

// agentInstallCmd represents the agent install command
var agentInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Fixpanic agent",
	Long: `Install the Fixpanic agent on your server.
	
This command downloads and installs the connectivity layer binary, creates the
necessary configuration files, and sets up the systemd service for automatic
startup.`,
	Example: `  # Install with agent credentials
  fixpanic agent install --agent-id="agent_123" --api-key="fp_abc123xyz"
  
  # Install with custom socket server
  fixpanic agent install --agent-id="agent_123" --api-key="fp_abc123xyz" --socket-server="custom.socket.com:8080"
  
  # Force reinstall
  fixpanic agent install --agent-id="agent_123" --api-key="fp_abc123xyz" --force`,
	RunE: runAgentInstall,
}

func init() {
	agentCmd.AddCommand(agentInstallCmd)

	// Add flags
	agentInstallCmd.Flags().StringVar(&agentID, "agent-id", "", "Agent ID from Fixpanic dashboard (required)")
	agentInstallCmd.Flags().StringVar(&agentAPIKey, "api-key", "", "Agent API key from Fixpanic dashboard (required)")
	agentInstallCmd.Flags().StringVar(&socketServer, "socket-server", "", "Socket server address (overrides config)")
	agentInstallCmd.Flags().BoolVar(&forceInstall, "force", false, "Force reinstall even if agent is already installed")

	// Mark required flags
	agentInstallCmd.MarkFlagRequired("agent-id")
	agentInstallCmd.MarkFlagRequired("api-key")
}

func runAgentInstall(cmd *cobra.Command, args []string) error {
	fmt.Println("Installing Fixpanic agent...")

	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if running as root for system-wide installation
	if !platformInfo.IsRoot {
		fmt.Println("Warning: Running as non-root user. Agent will be installed in user directories.")
		fmt.Printf("Binary location: %s\n", platformInfo.LibDir)
		fmt.Printf("Config location: %s\n", platformInfo.ConfigDir)
	}

	// Create necessary directories
	if err := platformInfo.CreateDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Check if already installed
	connectivityManager := connectivity.NewManager(platformInfo)
	if connectivityManager.IsInstalled() && !forceInstall {
		return fmt.Errorf("agent is already installed. Use --force to reinstall")
	}

	// Download connectivity layer
	fmt.Println("Downloading connectivity layer...")
	if err := connectivityManager.Download("latest"); err != nil {
		return fmt.Errorf("failed to download connectivity layer: %w", err)
	}

	// Create configuration
	agentConfig := config.DefaultConfig()
	agentConfig.Agent.ID = agentID
	agentConfig.Agent.APIKey = agentAPIKey

	// Use socket server from flag if provided, otherwise use default or config
	if socketServer != "" {
		agentConfig.Agent.SocketServer = socketServer
	} else if viper.GetString("socket_server") != "" {
		agentConfig.Agent.SocketServer = viper.GetString("socket_server")
	}

	// Validate configuration
	if err := agentConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Save configuration
	configPath := platformInfo.GetConfigPath()
	if err := config.SaveConfig(agentConfig, configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Configuration saved to: %s\n", configPath)

	// Install systemd service if available
	if platform.IsSystemdAvailable() {
		serviceManager := service.NewManager(platformInfo)

		// Remove old service if it exists
		if err := serviceManager.Uninstall(); err != nil {
			fmt.Printf("Warning: failed to remove old service: %v\n", err)
		}

		// Install new service
		if err := serviceManager.Install(); err != nil {
			fmt.Printf("Warning: failed to install systemd service: %v\n", err)
			fmt.Println("You can start the agent manually with: fixpanic agent start")
		} else {
			// Enable and start the service
			if err := serviceManager.Enable(); err != nil {
				fmt.Printf("Warning: failed to enable service: %v\n", err)
			}

			if err := serviceManager.Start(); err != nil {
				fmt.Printf("Warning: failed to start service: %v\n", err)
				fmt.Println("You can start the agent manually with: fixpanic agent start")
			} else {
				fmt.Println("Agent service installed and started successfully")
			}
		}
	} else {
		fmt.Println("Systemd not available. You can start the agent manually with: fixpanic agent start")
	}

	fmt.Println("\n✅ Fixpanic agent installed successfully!")
	fmt.Printf("Agent ID: %s\n", agentID)
	fmt.Printf("Binary location: %s\n", platformInfo.GetBinaryPath())
	fmt.Printf("Config location: %s\n", configPath)

	if platform.IsSystemdAvailable() {
		fmt.Println("\nThe agent will start automatically on system boot.")
		fmt.Println("You can manage the service with:")
		fmt.Printf("  sudo systemctl status %s\n", platform.GetSystemdServiceName())
		fmt.Printf("  sudo systemctl stop %s\n", platform.GetSystemdServiceName())
		fmt.Printf("  sudo systemctl restart %s\n", platform.GetSystemdServiceName())
	}

	return nil
}
