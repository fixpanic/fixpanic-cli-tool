package platform

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

// PlatformInfo contains platform-specific information
type PlatformInfo struct {
	OS        string
	Arch      string
	LibDir    string
	BinDir    string
	ConfigDir string
	LogDir    string
	IsRoot    bool
}

// GetPlatformInfo returns platform-specific information
func GetPlatformInfo() (*PlatformInfo, error) {
	os := runtime.GOOS
	arch := runtime.GOARCH
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	isRoot := currentUser.Uid == "0"

	var libDir, binDir, configDir, logDir string

	if isRoot {
		libDir = "/usr/local/lib/fixpanic"
		binDir = "/usr/local/bin"
		configDir = "/etc/fixpanic"
		logDir = "/var/log/fixpanic"
	} else {
		home := currentUser.HomeDir
		libDir = fmt.Sprintf("%s/.local/lib/fixpanic", home)
		binDir = fmt.Sprintf("%s/.local/bin", home)
		configDir = fmt.Sprintf("%s/.config/fixpanic", home)
		logDir = fmt.Sprintf("%s/.local/log/fixpanic", home)
	}

	return &PlatformInfo{
		OS:        os,
		Arch:      arch,
		LibDir:    libDir,
		BinDir:    binDir,
		ConfigDir: configDir,
		LogDir:    logDir,
		IsRoot:    isRoot,
	}, nil
}

// GetConnectivityBinaryName returns the connectivity binary name for the current platform
func GetConnectivityBinaryName() string {
	os := runtime.GOOS
	if os == "windows" {
		return "connectivity.exe"
	}
	return "connectivity"
}

// GetConnectivityDownloadURL returns the download URL for the connectivity binary
func GetConnectivityDownloadURL(version string) string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go arch names to our release names
	archMap := map[string]string{
		"amd64": "amd64",
		"arm64": "arm64",
		"386":   "386",
		"arm":   "arm",
	}

	releaseArch, ok := archMap[arch]
	if !ok {
		releaseArch = arch
	}

	baseURL := "https://releases.fixpanic.com/connectivity"
	if version == "latest" {
		return fmt.Sprintf("%s/latest/connectivity-%s-%s", baseURL, os, releaseArch)
	}
	return fmt.Sprintf("%s/%s/connectivity-%s-%s", baseURL, version, os, releaseArch)
}

// IsCommandAvailable checks if a command is available in PATH
func IsCommandAvailable(name string) bool {
	cmd := exec.Command("which", name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// IsSystemdAvailable checks if systemd is available on the system
func IsSystemdAvailable() bool {
	return IsCommandAvailable("systemctl")
}

// GetSystemdServiceName returns the systemd service name
func GetSystemdServiceName() string {
	return "fixpanic-agent.service"
}

// CreateDirectories creates the necessary directories for the agent
func (p *PlatformInfo) CreateDirectories() error {
	dirs := []string{
		p.LibDir,
		p.ConfigDir,
		p.LogDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetBinaryPath returns the full path to the connectivity binary
func (p *PlatformInfo) GetBinaryPath() string {
	return fmt.Sprintf("%s/%s", p.LibDir, GetConnectivityBinaryName())
}

// GetConfigPath returns the full path to the agent config file
func (p *PlatformInfo) GetConfigPath() string {
	return fmt.Sprintf("%s/agent.yaml", p.ConfigDir)
}

// GetServiceFilePath returns the full path to the systemd service file
func (p *PlatformInfo) GetServiceFilePath() string {
	return fmt.Sprintf("/etc/systemd/system/%s", GetSystemdServiceName())
}

// NormalizeArch normalizes architecture names for consistency
func NormalizeArch(arch string) string {
	arch = strings.ToLower(arch)
	switch arch {
	case "x86_64", "amd64":
		return "amd64"
	case "aarch64", "arm64":
		return "arm64"
	case "i386", "i686":
		return "386"
	case "armv7", "armv7l":
		return "arm"
	default:
		return arch
	}
}
