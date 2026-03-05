package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
)

func setupTestEnv(t *testing.T) (string, *platform.PlatformInfo) {
	tmpDir := t.TempDir()
	
	p := &platform.PlatformInfo{
		OS:        "linux",
		Arch:      "amd64",
		LibDir:    filepath.Join(tmpDir, "lib"),
		BinDir:    filepath.Join(tmpDir, "bin"),
		ConfigDir: filepath.Join(tmpDir, "etc"),
		LogDir:    filepath.Join(tmpDir, "log"),
		HomeDir:   filepath.Join(tmpDir, "home"),
		IsRoot:    false,
	}

	// Create directories
	_ = os.MkdirAll(p.LibDir, 0755)
	_ = os.MkdirAll(p.ConfigDir, 0755)
	_ = os.MkdirAll(p.LogDir, 0755)

	return tmpDir, p
}

func TestNodeStatus_NotInstalled(t *testing.T) {
	_, p := setupTestEnv(t)

	origGetPlatformInfo := platform.GetPlatformInfo
	platform.GetPlatformInfo = func() (*platform.PlatformInfo, error) {
		return p, nil
	}
	t.Cleanup(func() { platform.GetPlatformInfo = origGetPlatformInfo })

	rootCmd.SetArgs([]string{"node", "status"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestNodeStatus_Installed(t *testing.T) {
	_, p := setupTestEnv(t)

	origGetPlatformInfo := platform.GetPlatformInfo
	platform.GetPlatformInfo = func() (*platform.PlatformInfo, error) {
		return p, nil
	}
	t.Cleanup(func() { platform.GetPlatformInfo = origGetPlatformInfo })

	// Pre-install binary and config
	binaryPath := filepath.Join(p.LibDir, "opssquad-connectivity-layer")
	_ = os.WriteFile(binaryPath, []byte("fake binary"), 0755)
	
	configPath := filepath.Join(p.ConfigDir, "node.yaml")
	configContent := `app:
  node_id: test-node
  token: test-token
`
	_ = os.WriteFile(configPath, []byte(configContent), 0644)

	rootCmd.SetArgs([]string{"node", "status"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestNodeStart(t *testing.T) {
	_, p := setupTestEnv(t)

	origGetPlatformInfo := platform.GetPlatformInfo
	platform.GetPlatformInfo = func() (*platform.PlatformInfo, error) {
		return p, nil
	}
	t.Cleanup(func() { platform.GetPlatformInfo = origGetPlatformInfo })

	// Mock systemd as unavailable (so it tries background process mode)
	origIsSystemdAvailable := platform.IsSystemdAvailable
	platform.IsSystemdAvailable = func() bool { return false }
	t.Cleanup(func() { platform.IsSystemdAvailable = origIsSystemdAvailable })

	// Pre-install binary and config
	binaryPath := filepath.Join(p.LibDir, "opssquad-connectivity-layer")
	_ = os.WriteFile(binaryPath, []byte("#!/bin/sh\nsleep 10 &"), 0755)
	
	configPath := filepath.Join(p.ConfigDir, "node.yaml")
	_ = os.WriteFile(configPath, []byte("app:\n  node_id: test\n  token: test\n"), 0644)

	rootCmd.SetArgs([]string{"node", "start"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestNodeStop(t *testing.T) {
	_, p := setupTestEnv(t)

	origGetPlatformInfo := platform.GetPlatformInfo
	platform.GetPlatformInfo = func() (*platform.PlatformInfo, error) {
		return p, nil
	}
	t.Cleanup(func() { platform.GetPlatformInfo = origGetPlatformInfo })

	// Mock systemd as unavailable
	origIsSystemdAvailable := platform.IsSystemdAvailable
	platform.IsSystemdAvailable = func() bool { return false }
	t.Cleanup(func() { platform.IsSystemdAvailable = origIsSystemdAvailable })

	rootCmd.SetArgs([]string{"node", "stop"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}
