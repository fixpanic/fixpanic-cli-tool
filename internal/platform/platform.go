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

// GetFixPanicAgentBinaryName returns the correct binary name for FixPanic Agent
func GetFixPanicAgentBinaryName() string {
	if runtime.GOOS == "windows" {
		return "fixpanic-connectivity-layer.exe"
	}
	return "fixpanic-connectivity-layer"
}

// GetConnectivityBinaryName returns the connectivity binary name for the current platform (DEPRECATED)
// TODO: Remove this function after migration to GetFixPanicAgentBinaryName
func GetConnectivityBinaryName() string {
	fmt.Println("WARNING: GetConnectivityBinaryName is deprecated, use GetFixPanicAgentBinaryName instead")
	return GetFixPanicAgentBinaryName()
}

// GetFixPanicAgentBinaryPath returns the path to the FixPanic Agent binary
func (p *PlatformInfo) GetFixPanicAgentBinaryPath() string {
	return fmt.Sprintf("%s/%s", p.LibDir, GetFixPanicAgentBinaryName())
}

// GetFixPanicAgentPlatformInfo returns normalized platform info matching task requirements
func GetFixPanicAgentPlatformInfo() (os, arch string, err error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Normalize OS names as per task prompt
	switch goos {
	case "linux":
		os = "linux"
	case "darwin":
		os = "darwin"
	case "windows":
		os = "windows"
	default:
		return "", "", fmt.Errorf("unsupported operating system: %s", goos)
	}

	// Normalize architecture names (x86_64 -> amd64 as per task prompt)
	switch goarch {
	case "amd64", "x86_64":
		arch = "amd64"
	case "arm64", "aarch64":
		arch = "arm64"
	case "386", "i386", "i686":
		arch = "386"
	case "arm", "armv7":
		arch = "arm"
	default:
		return "", "", fmt.Errorf("unsupported architecture: %s", goarch)
	}

	return os, arch, nil
}

// GetFixPanicAgentDownloadURL returns the correct GitHub Releases URL
func GetFixPanicAgentDownloadURL(version string) (string, error) {
	os, arch, err := GetFixPanicAgentPlatformInfo()
	if err != nil {
		return "", fmt.Errorf("failed to get platform info: %w", err)
	}

	// Construct URL as per task prompt requirements
	baseURL := "https://github.com/fixpanic/fixpanic-connectivity-layer-release/releases"

	if version == "latest" {
		return fmt.Sprintf("%s/latest/download/fixpanic-connectivity-layer-%s-%s", baseURL, os, arch), nil
	}

	return fmt.Sprintf("%s/download/%s/fixpanic-connectivity-layer-%s-%s", baseURL, version, os, arch), nil
}

// GetConnectivityDownloadURL returns the download URL for the connectivity binary (DEPRECATED)
// TODO: Remove this function after migration to GetFixPanicAgentDownloadURL
func GetConnectivityDownloadURL(version string) string {
	fmt.Println("WARNING: GetConnectivityDownloadURL is deprecated, use GetFixPanicAgentDownloadURL instead")
	url, err := GetFixPanicAgentDownloadURL(version)
	if err != nil {
		// For backward compatibility, return empty string on error
		fmt.Printf("Error getting download URL: %v\n", err)
		return ""
	}
	return url
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
	return "fixpanic-connectivity-layer.service"
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
	return fmt.Sprintf("%s/%s", p.LibDir, GetFixPanicAgentBinaryName())
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
