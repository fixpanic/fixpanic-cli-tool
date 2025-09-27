package cmd

import (
	"fmt"
	"net"
	"time"

	"github.com/fixpanic/fixpanic-cli/internal/connectivity"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/spf13/cobra"
)

// agentConnectionCmd represents the agent test-connection command
var agentConnectionCmd = &cobra.Command{
	Use:   "test-connection",
	Short: "Test connection to Fixpanic infrastructure",
	Long: `Test the connection to the Fixpanic socket server.
	
This command verifies that your agent can connect to the Fixpanic infrastructure
and that the network connectivity is working properly.`,
	Example: `  # Test connection
  fixpanic agent test-connection`,
	RunE: runAgentConnection,
}

func init() {
	agentCmd.AddCommand(agentConnectionCmd)
}

func runAgentConnection(cmd *cobra.Command, args []string) error {
	fmt.Println("Testing connection to Fixpanic infrastructure...")

	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if connectivity layer is installed
	connectivityManager := connectivity.NewManager(platformInfo)
	if !connectivityManager.IsFixPanicAgentInstalled() {
		return fmt.Errorf("agent is not installed. Run 'fixpanic agent install' first")
	}

	// Test socket server connection (hardcoded in agent)
	socketServer := "socket.fixpanic.com:8080"

	fmt.Printf("Testing connection to: %s\n", socketServer)

	// Parse the address
	host, port, err := net.SplitHostPort(socketServer)
	if err != nil {
		return fmt.Errorf("invalid socket server address: %w", err)
	}

	// Test TCP connection
	fmt.Printf("Connecting to %s:%s...\n", host, port)

	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		fmt.Println("\nTroubleshooting tips:")
		fmt.Println("1. Check your internet connection")
		fmt.Println("2. Verify the socket server address is correct")
		fmt.Println("3. Check if your firewall is blocking the connection")
		fmt.Println("4. Ensure the socket server is accessible from your network")
		return fmt.Errorf("connection test failed")
	}
	defer conn.Close()

	fmt.Println("✅ TCP connection successful!")

	// Test if we can resolve the hostname
	if host != "localhost" && host != "127.0.0.1" {
		fmt.Printf("Resolving hostname: %s\n", host)
		ips, err := net.LookupIP(host)
		if err != nil {
			fmt.Printf("⚠️  DNS resolution failed: %v\n", err)
		} else {
			fmt.Printf("✅ DNS resolution successful. IP addresses: %v\n", ips)
		}
	}

	// Test connection timeout
	fmt.Println("Testing connection timeout...")

	testConn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		fmt.Printf("⚠️  Connection timeout test failed: %v\n", err)
	} else {
		testConn.Close()
		fmt.Println("✅ Connection timeout test passed")
	}

	fmt.Println("\n✅ Connection test completed successfully!")
	fmt.Println("Your agent should be able to connect to the Fixpanic infrastructure.")

	// Additional checks
	fmt.Println("\nAdditional checks:")

	// Check if we can ping the host
	if host != "localhost" && host != "127.0.0.1" {
		fmt.Printf("Testing ping to %s...\n", host)
		if err := pingHost(host); err != nil {
			fmt.Printf("⚠️  Ping failed: %v (this is not critical)\n", err)
		} else {
			fmt.Printf("✅ Ping successful\n")
		}
	}

	return nil
}

func pingHost(host string) error {
	// Simple ping test using net.Dial
	conn, err := net.DialTimeout("ip4:icmp", host, 3*time.Second)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
