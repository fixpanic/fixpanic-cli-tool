//go:build linux || freebsd || openbsd || netbsd
// +build linux freebsd openbsd netbsd

package process

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// UnixProcessManager handles process management on Unix-like systems (Linux, BSD)
type UnixProcessManager struct {
	BaseProcessManager
}

// NewUnixProcessManager creates a new Unix process manager
func NewUnixProcessManager() *UnixProcessManager {
	return &UnixProcessManager{}
}

// newPlatformProcessManager creates a new Unix process manager for the factory
func newPlatformProcessManager() ProcessManager {
	return NewUnixProcessManager()
}

// StartProcess starts a process on Unix-like systems with proper detachment
func (u *UnixProcessManager) StartProcess(config ProcessConfig) (*ProcessInfo, error) {
	cmd := exec.Command(config.BinaryPath, config.Args...)

	if config.WorkingDir != "" {
		cmd.Dir = config.WorkingDir
	}

	if len(config.Env) > 0 {
		cmd.Env = append(os.Environ(), config.Env...)
	}

	// Unix-specific process creation attributes
	if config.Detach {
		// Create new session and process group for proper detachment
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid:  true, // Create new session
			Setpgid: true, // Set process group ID
		}
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process on Unix: %w", err)
	}

	// Release the process to allow it to continue running independently
	if err := cmd.Process.Release(); err != nil {
		return nil, fmt.Errorf("failed to release process on Unix: %w", err)
	}

	return &ProcessInfo{
		PID:     cmd.Process.Pid,
		Running: true,
		Error:   nil,
	}, nil
}

// StopProcess stops a process on Unix-like systems
func (u *UnixProcessManager) StopProcess(pid int) error {
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

// UnixServiceManager provides systemd integration for Linux
type UnixServiceManager struct {
	BaseProcessManager
	serviceName string
}

// NewUnixServiceManager creates a new Unix systemd service manager
func NewUnixServiceManager(serviceName string) *UnixServiceManager {
	return &UnixServiceManager{
		serviceName: serviceName,
	}
}

// InstallService installs the agent as a systemd service
func (u *UnixServiceManager) InstallService(binaryPath, configPath string) error {
	// Generate systemd service file content
	serviceContent := u.generateServiceContent(binaryPath, configPath)

	// Get the systemd service file path
	servicePath := u.getServicePath()

	// Create the systemd directory if it doesn't exist
	systemdDir := "/etc/systemd/system"
	if err := os.MkdirAll(systemdDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd directory: %w", err)
	}

	// Write the service file
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write systemd service file: %w", err)
	}

	// Reload systemd
	cmd := exec.Command("systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	return nil
}

// StartService starts the systemd service
func (u *UnixServiceManager) StartService() error {
	cmd := exec.Command("systemctl", "start", u.serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start systemd service: %w", err)
	}

	return nil
}

// StopService stops the systemd service
func (u *UnixServiceManager) StopService() error {
	cmd := exec.Command("systemctl", "stop", u.serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop systemd service: %w", err)
	}

	return nil
}

// EnableService enables the systemd service to start on boot
func (u *UnixServiceManager) EnableService() error {
	cmd := exec.Command("systemctl", "enable", u.serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable systemd service: %w", err)
	}

	return nil
}

// DisableService disables the systemd service
func (u *UnixServiceManager) DisableService() error {
	cmd := exec.Command("systemctl", "disable", u.serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable systemd service: %w", err)
	}

	return nil
}

// GetServiceStatus gets the status of the systemd service
func (u *UnixServiceManager) GetServiceStatus() (string, error) {
	cmd := exec.Command("systemctl", "is-active", u.serviceName)
	output, err := cmd.Output()
	if err != nil {
		// Service is not active
		return "inactive", nil
	}

	status := string(output)
	if status == "active\n" {
		return "active", nil
	}

	return "inactive", nil
}

// IsServiceEnabled checks if the systemd service is enabled
func (u *UnixServiceManager) IsServiceEnabled() (bool, error) {
	cmd := exec.Command("systemctl", "is-enabled", u.serviceName)
	output, err := cmd.Output()
	if err != nil {
		return false, nil
	}

	return string(output) == "enabled\n", nil
}

// getServicePath returns the path to the systemd service file
func (u *UnixServiceManager) getServicePath() string {
	return "/etc/systemd/system/" + u.serviceName + ".service"
}

// generateServiceContent generates the systemd service file content
func (u *UnixServiceManager) generateServiceContent(binaryPath, configPath string) string {
	return fmt.Sprintf(`[Unit]
Description=FixPanic Agent - TCP socket connectivity layer for secure command execution
After=network.target

[Service]
Type=simple
ExecStart=%s --config %s
Restart=always
RestartSec=5
User=root
Group=root

[Install]
WantedBy=multi-user.target
`, binaryPath, configPath)
}

// Unix-specific helper functions

// GetUnixProcessInfo gets detailed process information on Unix-like systems
func GetUnixProcessInfo(pid int) (map[string]string, error) {
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

// CheckUnixProcessExists checks if a process exists on Unix-like systems
func CheckUnixProcessExists(pid int) bool {
	if pid <= 0 {
		return false
	}

	// Try to find the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// GetUnixProcessList gets a list of running processes on Unix-like systems
func GetUnixProcessList() ([]ProcessInfo, error) {
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get process list: %w", err)
	}

	lines := splitLines(string(output))
	var processes []ProcessInfo

	// Skip header line
	for i := 1; i < len(lines); i++ {
		fields := splitFields(lines[i])
		if len(fields) >= 11 {
			// Parse PID from the second field (ps aux format)
			var pid int
			if _, err := fmt.Sscanf(fields[1], "%d", &pid); err == nil {
				processes = append(processes, ProcessInfo{
					PID:     pid,
					Running: true,
					Error:   nil,
				})
			}
		}
	}

	return processes, nil
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
