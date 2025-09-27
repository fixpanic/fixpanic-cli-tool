//go:build darwin
// +build darwin

package process

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// DarwinProcessManager handles process management on macOS
type DarwinProcessManager struct {
	BaseProcessManager
}

// NewDarwinProcessManager creates a new Darwin process manager
func NewDarwinProcessManager() *DarwinProcessManager {
	return &DarwinProcessManager{}
}

// newPlatformProcessManager creates a new Darwin process manager for the factory
func newPlatformProcessManager() ProcessManager {
	return NewDarwinProcessManager()
}

// StartProcess starts a process on macOS with proper detachment
func (d *DarwinProcessManager) StartProcess(config ProcessConfig) (*ProcessInfo, error) {
	cmd := exec.Command(config.BinaryPath, config.Args...)

	if config.WorkingDir != "" {
		cmd.Dir = config.WorkingDir
	}

	if len(config.Env) > 0 {
		cmd.Env = append(os.Environ(), config.Env...)
	}

	// macOS-specific process creation attributes
	if config.Detach {
		// Use Unix-specific session creation for proper detachment
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid:  true, // Create new session
			Setpgid: true, // Set process group ID
		}
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process on macOS: %w", err)
	}

	// Release the process to allow it to continue running independently
	if err := cmd.Process.Release(); err != nil {
		return nil, fmt.Errorf("failed to release process on macOS: %w", err)
	}

	return &ProcessInfo{
		PID:     cmd.Process.Pid,
		Running: true,
		Error:   nil,
	}, nil
}

// StopProcess stops a process on macOS
func (d *DarwinProcessManager) StopProcess(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid PID: %d", pid)
	}

	// Try graceful termination first
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	// Send SIGTERM for graceful shutdown
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// If SIGTERM fails, try SIGKILL
		if err := process.Signal(syscall.SIGKILL); err != nil {
			return fmt.Errorf("failed to terminate process %d: %w", pid, err)
		}
	}

	return nil
}

// DarwinServiceManager provides macOS launchd integration
type DarwinServiceManager struct {
	BaseProcessManager
	serviceName string
}

// NewDarwinServiceManager creates a new macOS launchd service manager
func NewDarwinServiceManager(serviceName string) *DarwinServiceManager {
	return &DarwinServiceManager{
		serviceName: serviceName,
	}
}

// InstallService installs the agent as a macOS launchd service
func (d *DarwinServiceManager) InstallService(binaryPath, configPath string) error {
	// Generate launchd plist content
	plistContent := d.generatePlistContent(binaryPath, configPath)

	// Get the launchd plist path
	plistPath := d.getPlistPath()

	// Create the plist directory if it doesn't exist
	plistDir := "/Users/" + os.Getenv("USER") + "/Library/LaunchAgents"
	if err := os.MkdirAll(plistDir, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	// Write the plist file
	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	return nil
}

// StartService loads and starts the launchd service
func (d *DarwinServiceManager) StartService() error {
	plistPath := d.getPlistPath()

	// Load the service
	cmd := exec.Command("launchctl", "load", plistPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to load launchd service: %w", err)
	}

	// Start the service
	cmd = exec.Command("launchctl", "start", d.serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start launchd service: %w", err)
	}

	return nil
}

// StopService stops the launchd service
func (d *DarwinServiceManager) StopService() error {
	// Stop the service
	cmd := exec.Command("launchctl", "stop", d.serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop launchd service: %w", err)
	}

	return nil
}

// UnloadService unloads the launchd service
func (d *DarwinServiceManager) UnloadService() error {
	plistPath := d.getPlistPath()

	// Unload the service
	cmd := exec.Command("launchctl", "unload", plistPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unload launchd service: %w", err)
	}

	return nil
}

// GetServiceStatus gets the status of the launchd service
func (d *DarwinServiceManager) GetServiceStatus() (string, error) {
	// Check if service is loaded
	cmd := exec.Command("launchctl", "list")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list launchd services: %w", err)
	}

	outputStr := string(output)
	if contains(outputStr, d.serviceName) {
		// Service is loaded, check if it's running
		cmd = exec.Command("launchctl", "list", d.serviceName)
		output, err = cmd.Output()
		if err != nil {
			return "loaded", nil // Service is loaded but not running
		}

		// Parse the output to check status
		// launchctl list output format: PID Status Label
		if len(output) > 0 {
			return "running", nil
		}
	}

	return "not_loaded", nil
}

// getPlistPath returns the path to the launchd plist file
func (d *DarwinServiceManager) getPlistPath() string {
	return "/Users/" + os.Getenv("USER") + "/Library/LaunchAgents/" + d.serviceName + ".plist"
}

// generatePlistContent generates the launchd plist content
func (d *DarwinServiceManager) generatePlistContent(binaryPath, configPath string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>--config</string>
        <string>%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/fixpanic-agent.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/fixpanic-agent-error.log</string>
</dict>
</plist>`, d.serviceName, binaryPath, configPath)
}

// Darwin-specific helper functions

// GetDarwinProcessInfo gets detailed process information on macOS
func GetDarwinProcessInfo(pid int) (map[string]string, error) {
	// Use ps command to get process info
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "pid,ppid,command")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get process info: %w", err)
	}

	// Parse the output
	lines := splitLines(string(output))
	if len(lines) < 2 {
		return nil, fmt.Errorf("no process information found")
	}

	// Parse the header and data lines
	fields := splitFields(lines[1])
	if len(fields) < 3 {
		return nil, fmt.Errorf("invalid process information format")
	}

	return map[string]string{
		"pid":     fields[0],
		"ppid":    fields[1],
		"command": fields[2],
	}, nil
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	var currentLine []rune

	for _, r := range s {
		if r == '\n' {
			if len(currentLine) > 0 {
				lines = append(lines, string(currentLine))
				currentLine = []rune{}
			}
		} else {
			currentLine = append(currentLine, r)
		}
	}

	if len(currentLine) > 0 {
		lines = append(lines, string(currentLine))
	}

	return lines
}

// splitFields splits a line into fields
func splitFields(line string) []string {
	var fields []string
	var currentField []rune
	inField := false

	for _, r := range line {
		if r == ' ' || r == '\t' {
			if inField {
				fields = append(fields, string(currentField))
				currentField = []rune{}
				inField = false
			}
		} else {
			currentField = append(currentField, r)
			inField = true
		}
	}

	if len(currentField) > 0 {
		fields = append(fields, string(currentField))
	}

	return fields
}
