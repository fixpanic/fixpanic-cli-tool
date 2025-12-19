package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	serviceDir := filepath.Dir(servicePath)

	// Create service directory if it doesn't exist (important for user mode: ~/.config/systemd/user)
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd service directory: %w", err)
	}

	// Create systemd service file
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd
	if err := m.reloadSystemd(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	mode := "User"
	if m.platform.IsRoot {
		mode = "System"
	}
	fmt.Printf("%s systemd service installed: %s\n", mode, platform.GetSystemdServiceName())
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

	args := append(m.platform.GetSystemdCommandFlags(), "start", platform.GetSystemdServiceName())
	cmd := exec.Command("systemctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start service (output: %s): %w", strings.TrimSpace(string(output)), err)
	}

	fmt.Printf("Service started: %s\n", platform.GetSystemdServiceName())
	return nil
}

// Stop stops the service
func (m *Manager) Stop() error {
	if !platform.IsSystemdAvailable() {
		return fmt.Errorf("systemd is not available on this system")
	}

	args := append(m.platform.GetSystemdCommandFlags(), "stop", platform.GetSystemdServiceName())
	cmd := exec.Command("systemctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop service (output: %s): %w", strings.TrimSpace(string(output)), err)
	}

	fmt.Printf("Service stopped: %s\n", platform.GetSystemdServiceName())
	return nil
}

// Status returns the service status
func (m *Manager) Status() (string, error) {
	if !platform.IsSystemdAvailable() {
		return "systemd not available", nil
	}

	args := append(m.platform.GetSystemdCommandFlags(), "is-active", platform.GetSystemdServiceName())
	cmd := exec.Command("systemctl", args...)
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

	args := append(m.platform.GetSystemdCommandFlags(), "is-enabled", platform.GetSystemdServiceName())
	cmd := exec.Command("systemctl", args...)
	// Here we just want boolean, output doesn't matter much unless we want to debug IsEnabled failures specifically
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

	args := append(m.platform.GetSystemdCommandFlags(), "enable", platform.GetSystemdServiceName())
	cmd := exec.Command("systemctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to enable service (output: %s): %w", strings.TrimSpace(string(output)), err)
	}

	fmt.Printf("Service enabled for auto-start: %s\n", platform.GetSystemdServiceName())
	return nil
}

// Disable disables the service from starting on boot
func (m *Manager) Disable() error {
	if !platform.IsSystemdAvailable() {
		return fmt.Errorf("systemd is not available on this system")
	}

	args := append(m.platform.GetSystemdCommandFlags(), "disable", platform.GetSystemdServiceName())
	cmd := exec.Command("systemctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to disable service (output: %s): %w", strings.TrimSpace(string(output)), err)
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
{{ if .IsRoot }}User=root{{ end }}
ExecStart={{ .BinaryPath }} --config {{ .ConfigPath }}
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy={{ .WantedBy }}
`

	var wantedBy string
	if m.platform.IsRoot {
		wantedBy = "multi-user.target"
	} else {
		wantedBy = "default.target"
	}

	data := struct {
		IsRoot     bool
		BinaryPath string
		ConfigPath string
		WantedBy   string
	}{
		IsRoot:     m.platform.IsRoot,
		BinaryPath: binaryPath,
		ConfigPath: configPath,
		WantedBy:   wantedBy,
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

// IsUsable checks if systemd is actually usable (can connect to bus)
func (m *Manager) IsUsable() bool {
	if !platform.IsSystemdAvailable() {
		return false
	}

	// Try to list units (lightweight check) to see if we can connect to the bus
	args := append(m.platform.GetSystemdCommandFlags(), "list-units", "--no-pager", "-n", "0")
	cmd := exec.Command("systemctl", args...)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// reloadSystemd reloads the systemd daemon
func (m *Manager) reloadSystemd() error {
	args := append(m.platform.GetSystemdCommandFlags(), "daemon-reload")
	cmd := exec.Command("systemctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reload systemd daemon (output: %s): %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// GetServiceLogs returns the service logs
func (m *Manager) GetServiceLogs(lines int) (string, error) {
	if !platform.IsSystemdAvailable() {
		return "", fmt.Errorf("systemd is not available on this system")
	}

	// journalctl command structure might need to be adjusted: journalctl --user ...

	var args []string
	if !m.platform.IsRoot {
		args = append(args, "--user")
	}
	args = append(args, "-u", platform.GetSystemdServiceName(), "-n", fmt.Sprintf("%d", lines), "--no-pager")

	cmd := exec.Command("journalctl", args...)
	output, err := cmd.CombinedOutput() // Capture stderr too just in case
	if err != nil {
		return "", fmt.Errorf("failed to get service logs (output: %s): %w", strings.TrimSpace(string(output)), err)
	}

	return string(output), nil
}
