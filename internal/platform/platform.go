package platform

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
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
	HomeDir   string
	IsRoot    bool
}

// GetPlatformInfo returns platform-specific information
var GetPlatformInfo = func() (*PlatformInfo, error) {
	osType := runtime.GOOS
	arch := runtime.GOARCH
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	isRoot := currentUser.Uid == "0"
	home := currentUser.HomeDir

	var libDir, binDir, configDir, logDir string

	if isRoot {
		libDir = "/usr/local/lib/opssquad"
		binDir = "/usr/local/bin"
		configDir = "/etc/opssquad"
		logDir = "/var/log/opssquad"
	} else {
		libDir = fmt.Sprintf("%s/.local/lib/opssquad", home)
		binDir = fmt.Sprintf("%s/.local/bin", home)
		configDir = fmt.Sprintf("%s/.config/opssquad", home)
		logDir = fmt.Sprintf("%s/.local/log/opssquad", home)
	}

	return &PlatformInfo{
		OS:        osType,
		Arch:      arch,
		LibDir:    libDir,
		BinDir:    binDir,
		ConfigDir: configDir,
		LogDir:    logDir,
		HomeDir:   home,
		IsRoot:    isRoot,
	}, nil
}

// GetBinaryName returns the correct binary name for OpsSquad Node
func GetBinaryName() string {
	if runtime.GOOS == "windows" {
		return "opssquad-connectivity-layer.exe"
	}
	return "opssquad-connectivity-layer"
}

// GetConnectivityBinaryName returns the connectivity binary name for the current platform (DEPRECATED)
// TODO: Remove this function after migration to GetBinaryName
func GetConnectivityBinaryName() string {
	fmt.Println("WARNING: GetConnectivityBinaryName is deprecated, use GetBinaryName instead")
	return GetBinaryName()
}

// GetBinaryPath returns the full path to the connectivity binary
func (p *PlatformInfo) GetBinaryPath() string {
	return fmt.Sprintf("%s/%s", p.LibDir, GetBinaryName())
}

// GetOpsSquadNodeBinaryPath returns the path to the OpsSquad Node binary (DEPRECATED)
// TODO: Remove this function after migration to GetBinaryPath
func (p *PlatformInfo) GetOpsSquadNodeBinaryPath() string {
	return p.GetBinaryPath()
}

// GetNodePlatformInfo returns normalized platform info matching task requirements
func GetNodePlatformInfo() (os, arch string, err error) {
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

// GetBinaryDownloadURL returns the correct GitHub Releases URL
func GetBinaryDownloadURL(version string) (string, error) {
	os, arch, err := GetNodePlatformInfo()
	if err != nil {
		return "", fmt.Errorf("failed to get platform info: %w", err)
	}

	// Construct URL as per task prompt requirements
	baseURL := "https://github.com/fixpanic/opssquad-connectivity-layer-release/releases"

	if version == "latest" {
		return fmt.Sprintf("%s/latest/download/opssquad-connectivity-layer-%s-%s", baseURL, os, arch), nil
	}

	return fmt.Sprintf("%s/download/%s/opssquad-connectivity-layer-%s-%s", baseURL, version, os, arch), nil
}

// GetConnectivityDownloadURL returns the download URL for the connectivity binary (DEPRECATED)
// TODO: Remove this function after migration to GetBinaryDownloadURL
func GetConnectivityDownloadURL(version string) string {
	fmt.Println("WARNING: GetConnectivityDownloadURL is deprecated, use GetBinaryDownloadURL instead")
	url, err := GetBinaryDownloadURL(version)
	if err != nil {
		// For backward compatibility, return empty string on error
		fmt.Printf("Error getting download URL: %v\n", err)
		return ""
	}
	return url
}

// IsCommandAvailable checks if a command is available in PATH
var IsCommandAvailable = func(name string) bool {
	cmd := exec.Command("which", name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// IsSystemdAvailable checks if systemd is available on the system
var IsSystemdAvailable = func() bool {
	return IsCommandAvailable("systemctl")
}

// GetSystemdServiceName returns the systemd service name
func GetSystemdServiceName() string {
	return "opssquad-connectivity-layer.service"
}

// CreateDirectories creates the necessary directories for the node
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

// GetConfigPath returns the full path to the node config file
func (p *PlatformInfo) GetConfigPath() string {
	return fmt.Sprintf("%s/node.yaml", p.ConfigDir)
}

// GetServiceFilePath returns the full path to the systemd service file
func (p *PlatformInfo) GetServiceFilePath() string {
	if p.IsRoot {
		return fmt.Sprintf("/etc/systemd/system/%s", GetSystemdServiceName())
	}
	// For user mode systemd, the standard path is ~/.config/systemd/user/
	return filepath.Join(p.HomeDir, ".config", "systemd", "user", GetSystemdServiceName())
}

// GetSystemdCommandFlags returns flags for systemctl commands based on permissions
func (p *PlatformInfo) GetSystemdCommandFlags() []string {
	if p.IsRoot {
		return []string{}
	}
	return []string{"--user"}
}

// AutoConfigureEnvironment attempts to fix common environment issues and returns warnings if they cannot be fixed
func (p *PlatformInfo) AutoConfigureEnvironment() []string {
	var warnings []string

	// Check if XDG_RUNTIME_DIR is set for user mode systemd
	if !p.IsRoot && runtime.GOOS == "linux" && IsSystemdAvailable() {
		if os.Getenv("XDG_RUNTIME_DIR") == "" {
			uid := os.Getuid()
			// Try to find the runtime directory
			potentialPath := fmt.Sprintf("/run/user/%d", uid)
			if _, err := os.Stat(potentialPath); err == nil {
				// Found it! Set the environment variable for this process
				os.Setenv("XDG_RUNTIME_DIR", potentialPath)
				// No warning needed, we fixed it transparently
			} else {
				// Could not find it, return warning
				warnings = append(warnings, fmt.Sprintf("XDG_RUNTIME_DIR is not set and could not be auto-detected. This is required for user-mode systemd.\n   Try running: export XDG_RUNTIME_DIR=%s", potentialPath))
			}
		}
	}

	return warnings
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
