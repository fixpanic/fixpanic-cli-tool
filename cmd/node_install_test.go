package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
)

func TestNodeInstall(t *testing.T) {
	// Create a temp directory for our test "system"
	tmpDir := t.TempDir()
	
	libDir := filepath.Join(tmpDir, "lib")
	configDir := filepath.Join(tmpDir, "etc")
	logDir := filepath.Join(tmpDir, "log")
	homeDir := filepath.Join(tmpDir, "home")

	// Mock platform info
	origGetPlatformInfo := platform.GetPlatformInfo
	platform.GetPlatformInfo = func() (*platform.PlatformInfo, error) {
		return &platform.PlatformInfo{
			OS:        "linux",
			Arch:      "amd64",
			LibDir:    libDir,
			ConfigDir: configDir,
			LogDir:    logDir,
			HomeDir:   homeDir,
			IsRoot:    false,
		}, nil
	}
	t.Cleanup(func() { platform.GetPlatformInfo = origGetPlatformInfo })

	// Mock systemd as unavailable for this test to avoid exec calls
	origIsSystemdAvailable := platform.IsSystemdAvailable
	platform.IsSystemdAvailable = func() bool { return false }
	t.Cleanup(func() { platform.IsSystemdAvailable = origIsSystemdAvailable })

	// We also need to mock IsNodeUpdateAvailable to avoid network calls 
	// This is harder because it's inside connectivity.Manager which is instantiated in runNodeInstall.
	// However, we refactored connectivity.Manager to use interfaces.
	// But runNodeInstall uses connectivity.NewManager(platformInfo) which returns a RealManager with RealHTTPClient.
	
	// For a true integration test without network, we would need to either:
	// 1. Mock the GitHub API using httptest and hope the URL can be overridden (it's currently hardcoded)
	// 2. Mock the Manager instantiation (requires more refactoring of the command)
	
	// Let's try to at least run the command and see it failing at the network step, 
	// or mock the network if possible.
	
	// The URL is hardcoded in manager.go: https://api.github.com/repos/...
	// We can't easily override it without more refactoring.
	
	// For now, let's just test that the flags are registered and the command can be called.
	
	rootCmd.SetArgs([]string{"node", "install", "--node-id", "test-id", "--token", "test-token", "--force"})
	
	// This will still attempt to download if it's not installed.
	// To make this test pass without network, we can pre-create the binary.
	
	err := os.MkdirAll(libDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create libDir: %v", err)
	}
	binaryPath := filepath.Join(libDir, "opssquad-connectivity-layer")
	err = os.WriteFile(binaryPath, []byte("fake binary"), 0755)
	if err != nil {
		t.Fatalf("Failed to create fake binary: %v", err)
	}

	// Now it might still try to check for updates.
	// If it fails to check for updates, it logs a warning and continues (I added this in EnsureLatestNode).
	
	err = rootCmd.Execute()
	
	// If there's no internet, GetLatestNodeVersion will fail, log a warning, 
	// and continue with the "fake binary" we created.
	
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Verify config was created
	configPath := filepath.Join(configDir, "node.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created at %s", configPath)
	}
}
