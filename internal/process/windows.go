//go:build windows
// +build windows

package process

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// WindowsProcessManager handles process management on Windows
type WindowsProcessManager struct {
	BaseProcessManager
}

// NewWindowsProcessManager creates a new Windows process manager
func NewWindowsProcessManager() *WindowsProcessManager {
	return &WindowsProcessManager{}
}

// newPlatformProcessManager creates a new Windows process manager for the factory
func newPlatformProcessManager() ProcessManager {
	return NewWindowsProcessManager()
}

// StartProcess starts a process on Windows with proper detachment
func (w *WindowsProcessManager) StartProcess(config ProcessConfig) (*ProcessInfo, error) {
	cmd := exec.Command(config.BinaryPath, config.Args...)

	if config.WorkingDir != "" {
		cmd.Dir = config.WorkingDir
	}

	if len(config.Env) > 0 {
		cmd.Env = append(os.Environ(), config.Env...)
	}

	// Windows-specific process creation attributes
	if config.Detach {
		// Use Windows-specific creation flags to detach the process
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow:    true,
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x00000008, // DETACHED_PROCESS
		}
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process on Windows: %w", err)
	}

	// Release the process to allow it to continue running independently
	if err := cmd.Process.Release(); err != nil {
		return nil, fmt.Errorf("failed to release process on Windows: %w", err)
	}

	return &ProcessInfo{
		PID:     cmd.Process.Pid,
		Running: true,
		Error:   nil,
	}, nil
}

// StopProcess stops a process on Windows
func (w *WindowsProcessManager) StopProcess(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID: %d", pid)
	}

	// Open the process with termination rights
	handle, err := syscall.OpenProcess(syscall.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return fmt.Errorf("failed to open process %d: %w", pid, err)
	}
	defer syscall.CloseHandle(handle)

	// Terminate the process
	if err := syscall.TerminateProcess(handle, 0); err != nil {
		return fmt.Errorf("failed to terminate process %d: %w", pid, err)
	}

	return nil
}

// Windows-specific helper functions

// GetProcessExitCode gets the exit code of a Windows process
func GetProcessExitCode(pid int) (uint32, error) {
	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return 0, fmt.Errorf("failed to open process %d: %w", pid, err)
	}
	defer syscall.CloseHandle(handle)

	var exitCode uint32
	if err := syscall.GetExitCodeProcess(handle, &exitCode); err != nil {
		return 0, fmt.Errorf("failed to get exit code for process %d: %w", pid, err)
	}

	return exitCode, nil
}

// IsProcessRunningWindows provides a more reliable Windows-specific process check
func IsProcessRunningWindows(pid int) bool {
	if pid <= 0 {
		return false
	}

	// Try to open the process
	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	// Get the exit code
	var exitCode uint32
	if err := syscall.GetExitCodeProcess(handle, &exitCode); err != nil {
		return false
	}

	// If exit code is STILL_ACTIVE (259), the process is running
	return exitCode == 259 // STILL_ACTIVE
}

// WindowsServiceManager provides Windows Service integration
type WindowsServiceManager struct {
	BaseProcessManager
	serviceName string
}

// NewWindowsServiceManager creates a new Windows service manager
func NewWindowsServiceManager(serviceName string) *WindowsServiceManager {
	return &WindowsServiceManager{
		serviceName: serviceName,
	}
}

// InstallService installs the agent as a Windows service
func (w *WindowsServiceManager) InstallService(binaryPath, configPath string) error {
	// Use sc.exe to create the service
	cmd := exec.Command("sc.exe", "create", w.serviceName,
		fmt.Sprintf("binPath=%s --config %s", binaryPath, configPath),
		"start=auto",
		"displayname=FixPanic Agent")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create Windows service: %w", err)
	}

	return nil
}

// StartService starts the Windows service
func (w *WindowsServiceManager) StartService() error {
	cmd := exec.Command("sc.exe", "start", w.serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start Windows service: %w", err)
	}
	return nil
}

// StopService stops the Windows service
func (w *WindowsServiceManager) StopService() error {
	cmd := exec.Command("sc.exe", "stop", w.serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop Windows service: %w", err)
	}
	return nil
}

// DeleteService removes the Windows service
func (w *WindowsServiceManager) DeleteService() error {
	cmd := exec.Command("sc.exe", "delete", w.serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete Windows service: %w", err)
	}
	return nil
}

// GetServiceStatus gets the status of the Windows service
func (w *WindowsServiceManager) GetServiceStatus() (string, error) {
	cmd := exec.Command("sc.exe", "query", w.serviceName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to query Windows service: %w", err)
	}

	outputStr := string(output)
	if contains(outputStr, "RUNNING") {
		return "running", nil
	} else if contains(outputStr, "STOPPED") {
		return "stopped", nil
	} else if contains(outputStr, "START_PENDING") {
		return "starting", nil
	} else if contains(outputStr, "STOP_PENDING") {
		return "stopping", nil
	}

	return "unknown", nil
}
