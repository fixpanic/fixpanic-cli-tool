package service

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"text/template"

	"github.com/fixpanic/fixpanic-cli/internal/platform"
)

// Manager handles systemd service operations
type Manager struct {
	platform *platform.PlatformInfo
}

// NewManager creates a new service manager
func NewManager(platform *platform.PlatformInfo) *Manager {
	return &Manager{
		platform: platform,
	}
}

// Install installs the systemd service
func (m *Manager) Install() error {
	if !platform.IsSystemdAvailable() {
		return fmt.Errorf("systemd is not available on this system")
	}

	serviceContent, err := m.generateServiceFile()
	if err != nil {
		return fmt.Errorf("failed to generate service file: %w", err)
	}

	servicePath := m.platform.GetServiceFilePath()

	// Create systemd service file
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd
	if err := m.reloadSystemd(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	fmt.Printf("Systemd service installed: %s\n", platform.GetSystemdServiceName())
	return nil
}

// Uninstall removes the systemd service
func (m *Manager) Uninstall() error {
	if !platform.IsSystemdAvailable() {
		return nil // Nothing to do if systemd is not available
	}

	// Stop the service first
	if err := m.Stop(); err != nil {
		// Continue even if stop fails
		fmt.Printf("Warning: failed to stop service: %v\n", err)
	}

	servicePath := m.platform.GetServiceFilePath()

	// Remove service file
	if err := os.Remove(servicePath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already removed
		}
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Reload systemd
	if err := m.reloadSystemd(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	fmt.Printf("Systemd service uninstalled: %s\n", platform.GetSystemdServiceName())
	return nil
}

// Start starts the service
func (m *Manager) Start() error {
	if !platform.IsSystemdAvailable() {
		return fmt.Errorf("systemd is not available on this system")
	}

	cmd := exec.Command("systemctl", "start", platform.GetSystemdServiceName())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	fmt.Printf("Service started: %s\n", platform.GetSystemdServiceName())
	return nil
}

// Stop stops the service
func (m *Manager) Stop() error {
	if !platform.IsSystemdAvailable() {
		return fmt.Errorf("systemd is not available on this system")
	}

	cmd := exec.Command("systemctl", "stop", platform.GetSystemdServiceName())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	fmt.Printf("Service stopped: %s\n", platform.GetSystemdServiceName())
	return nil
}

// Status returns the service status
func (m *Manager) Status() (string, error) {
	if !platform.IsSystemdAvailable() {
		return "systemd not available", nil
	}

	cmd := exec.Command("systemctl", "is-active", platform.GetSystemdServiceName())
	output, err := cmd.Output()
	if err != nil {
		// Service is not active
		return "inactive", nil
	}

	status := strings.TrimSpace(string(output))
	return status, nil
}

// IsEnabled checks if the service is enabled
func (m *Manager) IsEnabled() (bool, error) {
	if !platform.IsSystemdAvailable() {
		return false, nil
	}

	cmd := exec.Command("systemctl", "is-enabled", platform.GetSystemdServiceName())
	if err := cmd.Run(); err != nil {
		return false, nil // Service is not enabled
	}

	return true, nil
}

// Enable enables the service to start on boot
func (m *Manager) Enable() error {
	if !platform.IsSystemdAvailable() {
		return fmt.Errorf("systemd is not available on this system")
	}

	cmd := exec.Command("systemctl", "enable", platform.GetSystemdServiceName())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	fmt.Printf("Service enabled for auto-start: %s\n", platform.GetSystemdServiceName())
	return nil
}

// Disable disables the service from starting on boot
func (m *Manager) Disable() error {
	if !platform.IsSystemdAvailable() {
		return fmt.Errorf("systemd is not available on this system")
	}

	cmd := exec.Command("systemctl", "disable", platform.GetSystemdServiceName())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable service: %w", err)
	}

	fmt.Printf("Service disabled from auto-start: %s\n", platform.GetSystemdServiceName())
	return nil
}

// generateServiceFile generates the systemd service file content
func (m *Manager) generateServiceFile() (string, error) {
	binaryPath := m.platform.GetBinaryPath()
	configPath := m.platform.GetConfigPath()

	tmpl := `[Unit]
Description=Fixpanic Agent
After=network.target

[Service]
Type=simple
User={{ .User }}
ExecStart={{ .BinaryPath }} --config {{ .ConfigPath }}
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`

	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	user := currentUser.Username
	if m.platform.IsRoot {
		user = "root"
	}

	data := struct {
		User       string
		BinaryPath string
		ConfigPath string
	}{
		User:       user,
		BinaryPath: binaryPath,
		ConfigPath: configPath,
	}

	t, err := template.New("service").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	if err := t.Execute(&result, data); err != nil {
		return "", err
	}

	return result.String(), nil
}

// reloadSystemd reloads the systemd daemon
func (m *Manager) reloadSystemd() error {
	cmd := exec.Command("systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload systemd daemon: %w", err)
	}
	return nil
}

// GetServiceLogs returns the service logs
func (m *Manager) GetServiceLogs(lines int) (string, error) {
	if !platform.IsSystemdAvailable() {
		return "", fmt.Errorf("systemd is not available on this system")
	}

	args := []string{"journalctl", "-u", platform.GetSystemdServiceName(), "-n", fmt.Sprintf("%d", lines), "--no-pager"}
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get service logs: %w", err)
	}

	return string(output), nil
}
