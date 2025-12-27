package connectivity

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/fixpanic/opssquad-cli-tool/internal/logger"
	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
)

// Manager handles connectivity layer binary operations
type Manager struct {
	platform *platform.PlatformInfo
	client   *http.Client
}

// NewManager creates a new connectivity manager
func NewManager(platform *platform.PlatformInfo) *Manager {
	return &Manager{
		platform: platform,
		client:   &http.Client{},
	}
}

// Download downloads the connectivity layer binary
func (m *Manager) Download(version string) error {
	return m.DownloadBinary(version)
}

// DownloadBinary downloads the OpsSquad Node binary from GitHub Releases
func (m *Manager) DownloadBinary(version string) error {
	url, err := platform.GetBinaryDownloadURL(version)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}
	binaryPath := m.platform.GetBinaryPath()

	fmt.Printf("Downloading connectivity layer from %s...\n", url)

	// Create temporary file
	tmpFile := binaryPath + ".tmp"

	resp, err := m.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download binary: HTTP %d", resp.StatusCode)
	}

	// Create the file
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		out.Close()
		os.Remove(tmpFile)
		return fmt.Errorf("failed to save binary: %w", err)
	}

	// Sync to ensure all data is written to disk before closing
	if err := out.Sync(); err != nil {
		out.Close()
		os.Remove(tmpFile)
		return fmt.Errorf("failed to sync file to disk: %w", err)
	}

	// Close the file before chmod and rename
	if err := out.Close(); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Make the binary executable
	if err := os.Chmod(tmpFile, 0755); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// On macOS, remove quarantine attribute to allow execution
	if runtime.GOOS == "darwin" {
		if err := exec.Command("xattr", "-d", "com.apple.quarantine", tmpFile).Run(); err != nil {
			// Log warning but don't fail - quarantine removal is not critical
			logger.Warning("Failed to remove quarantine attribute: %v", err)
		}
	}

	// Move to final location
	if err := os.Rename(tmpFile, binaryPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to move binary to final location: %w", err)
	}

	fmt.Printf("Connectivity layer downloaded to %s\n", binaryPath)
	return nil
}

// IsInstalled checks if the connectivity layer is installed (DEPRECATED)
// TODO: Remove this function after migration to IsBinaryInstalled
func (m *Manager) IsInstalled() bool {
	fmt.Println("WARNING: IsInstalled() is deprecated, use IsBinaryInstalled() instead")
	return m.IsBinaryInstalled()
}

// GetVersion returns the version of the installed connectivity layer (DEPRECATED)
// TODO: Remove this function after migration to GetBinaryVersion
func (m *Manager) GetVersion() (string, error) {
	fmt.Println("WARNING: GetVersion() is deprecated, use GetBinaryVersion() instead")
	return m.GetBinaryVersion()
}

// Remove removes the connectivity layer binary (DEPRECATED)
// TODO: Remove this function after migration to RemoveBinary
func (m *Manager) Remove() error {
	fmt.Println("WARNING: Remove() is deprecated, use RemoveBinary() instead")
	return m.RemoveBinary()
}

// VerifyChecksum verifies the binary checksum
func (m *Manager) VerifyChecksum(expectedChecksum string) error {
	binaryPath := m.platform.GetBinaryPath()

	file, err := os.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to open binary: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualChecksum := fmt.Sprintf("%x", hash.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// GetBinaryPath returns the path to the connectivity binary
func (m *Manager) GetBinaryPath() string {
	return m.platform.GetBinaryPath()
}

// IsBinaryInstalled checks if the OpsSquad Node is installed
func (m *Manager) IsBinaryInstalled() bool {
	binaryPath := m.platform.GetBinaryPath()
	_, err := os.Stat(binaryPath)
	return err == nil
}

// GetBinaryVersion returns the version of the installed OpsSquad Node
func (m *Manager) GetBinaryVersion() (string, error) {
	binaryPath := m.platform.GetBinaryPath()

	if !m.IsBinaryInstalled() {
		return "", fmt.Errorf("OpsSquad Node not installed")
	}

	// Execute with --version flag
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// UpdateBinary updates the OpsSquad Node to the specified version
func (m *Manager) UpdateBinary(version string) error {
	fmt.Printf("Updating OpsSquad Node to version %s...\n", version)

	// Remove old version
	if err := m.RemoveBinary(); err != nil {
		return fmt.Errorf("failed to remove old version: %w", err)
	}

	// Download new version
	if err := m.DownloadBinary(version); err != nil {
		return fmt.Errorf("failed to download new version: %w", err)
	}

	fmt.Printf("OpsSquad Node updated successfully\n")
	return nil
}

// RemoveBinary removes the OpsSquad Node binary
func (m *Manager) RemoveBinary() error {
	binaryPath := m.platform.GetBinaryPath()

	if err := os.Remove(binaryPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already removed
		}
		return fmt.Errorf("failed to remove binary: %w", err)
	}

	return nil
}

// Update updates the connectivity layer to the specified version (DEPRECATED)
// TODO: Remove this function after migration to UpdateBinary
func (m *Manager) Update(version string) error {
	fmt.Println("WARNING: Update() is deprecated, use UpdateBinary() instead")
	return m.UpdateBinary(version)
}

// NodeRelease represents a GitHub release for the node binary
type NodeRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	PublishedAt string `json:"published_at"`
}

// GetLatestNodeVersion fetches the latest node version from GitHub releases
func (m *Manager) GetLatestNodeVersion() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	url := "https://api.github.com/repos/fixpanic/opssquad-connectivity-layer-release/releases/latest"

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API request failed: %d", resp.StatusCode)
	}

	var release NodeRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}

	return release.TagName, nil
}

// IsNodeUpdateAvailable checks if a newer version of the node is available
func (m *Manager) IsNodeUpdateAvailable() (bool, string, error) {
	if !m.IsBinaryInstalled() {
		return true, "", nil // Need to install
	}

	currentVersion, err := m.GetBinaryVersion()
	if err != nil {
		return true, "", fmt.Errorf("failed to get current version: %w", err)
	}

	latestVersion, err := m.GetLatestNodeVersion()
	if err != nil {
		return false, "", fmt.Errorf("failed to get latest version: %w", err)
	}

	// Parse version strings to compare them
	// For simplicity, we'll do string comparison since they follow semantic versioning
	currentClean := strings.TrimSpace(currentVersion)
	latestClean := strings.TrimSpace(latestVersion)

	// Extract version from output like "opssquad-connectivity-layer v1.0.0 - ..."
	if strings.Contains(currentClean, " v") {
		parts := strings.Split(currentClean, " v")
		if len(parts) > 1 {
			versionPart := strings.Split(parts[1], " ")[0]
			currentClean = "v" + versionPart
		}
	}

	return currentClean != latestClean, latestClean, nil
}

// EnsureLatestNode checks and updates the node binary if needed
func (m *Manager) EnsureLatestNode() error {
	logger.Progress("Checking for node binary updates")

	updateAvailable, latestVersion, err := m.IsNodeUpdateAvailable()
	if err != nil {
		logger.Warning("Failed to check for updates: %v", err)
		// Continue with existing binary if update check fails
		return nil
	}

	if !updateAvailable {
		if m.IsBinaryInstalled() {
			logger.List("Node binary is up to date")
		}
		return nil
	}

	// Update or install the node
	if m.IsBinaryInstalled() {
		currentVersion, _ := m.GetBinaryVersion()
		logger.Info("Node update available: %s → %s", currentVersion, latestVersion)
		logger.Progress("Downloading latest node binary")
	} else {
		logger.Progress("Installing node binary")
	}

	if err := m.DownloadBinary("latest"); err != nil {
		return fmt.Errorf("failed to download latest node: %w", err)
	}

	// Verify the update
	newVersion, err := m.GetBinaryVersion()
	if err != nil {
		logger.Warning("Failed to verify new version: %v", err)
	} else {
		logger.Success("Node binary updated to: %s", newVersion)
	}

	return nil
}
