// Package process provides cross-platform process management functionality
package process

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

// ProcessConfig contains configuration for starting a process
type ProcessConfig struct {
	BinaryPath string
	Args       []string
	WorkingDir string
	Env        []string
	Detach     bool
}

// ProcessInfo contains information about a running process
type ProcessInfo struct {
	PID     int
	Running bool
	Error   error
}

// ProcessManager provides cross-platform process management
type ProcessManager interface {
	StartProcess(config ProcessConfig) (*ProcessInfo, error)
	StopProcess(pid int) error
	IsProcessRunning(pid int) bool
	GetProcessStatus(pid int) *ProcessInfo
}

// NewProcessManager creates a platform-specific process manager
func NewProcessManager() ProcessManager {
	// This function will be implemented in platform-specific files
	// using build constraints to avoid cross-platform compilation issues
	return newPlatformProcessManager()
}

// BaseProcessManager provides common functionality
type BaseProcessManager struct{}

// GetProcessStatus checks if a process is running
func (b *BaseProcessManager) GetProcessStatus(pid int) *ProcessInfo {
	running := b.IsProcessRunning(pid)
	return &ProcessInfo{
		PID:     pid,
		Running: running,
		Error:   nil,
	}
}

// IsProcessRunning checks if a process with the given PID is running
func (b *BaseProcessManager) IsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	// Try to send signal 0 to check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix systems, signal 0 can be used to check if process exists
	if runtime.GOOS != "windows" {
		err = process.Signal(syscall.Signal(0))
		return err == nil
	}

	// On Windows, we need a different approach
	return b.isWindowsProcessRunning(pid)
}

// isWindowsProcessRunning checks if a Windows process is running
func (b *BaseProcessManager) isWindowsProcessRunning(pid int) bool {
	// Use tasklist command to check if process exists
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH", "/FO", "CSV")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Check if output contains the PID
	return len(output) > 0 && !isWindowsProcessNotFound(output)
}

// isWindowsProcessNotFound checks if tasklist output indicates process not found
func isWindowsProcessNotFound(output []byte) bool {
	outputStr := string(output)
	return contains(outputStr, "No tasks are running") ||
		contains(outputStr, "INFO: No tasks") ||
		contains(outputStr, "ERROR:")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
