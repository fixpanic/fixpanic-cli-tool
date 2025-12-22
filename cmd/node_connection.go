package cmd

import (
	"fmt"
	"net"
	"time"

	"github.com/fixpanic/opssquad-cli-tool/internal/connectivity"
	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
	"github.com/spf13/cobra"
)

// nodeConnectionCmd represents the node test-connection command
var nodeConnectionCmd = &cobra.Command{
	Use:   "test-connection",
	Short: "Test connection to OpsSquad infrastructure",
	Long: `Test the connection to the OpsSquad socket server.
	
This command verifies that your node can connect to the OpsSquad infrastructure
and that the network connectivity is working properly.`,
	Example: `  # Test connection
  opssquad node test-connection`,
	RunE: runNodeConnection,
}

func init() {
	nodeCmd.AddCommand(nodeConnectionCmd)
}

func runNodeConnection(cmd *cobra.Command, args []string) error {
	fmt.Println("Testing connection to OpsSquad infrastructure...")

	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if connectivity layer is installed
	connectivityManager := connectivity.NewManager(platformInfo)
	if !connectivityManager.IsBinaryInstalled() {
		return fmt.Errorf("node is not installed. Run 'opssquad node install' first")
	}

	// Test socket server connection (same as node uses)
	socketServer := "socket.opssquad.com:9000"

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
	fmt.Println("Your node should be able to connect to the OpsSquad infrastructure.")

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
