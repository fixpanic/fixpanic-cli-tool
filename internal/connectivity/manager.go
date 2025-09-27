package connectivity

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fixpanic/fixpanic-cli/internal/logger"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
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
	url := platform.GetConnectivityDownloadURL(version)
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
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to save binary: %w", err)
	}

	// Make the binary executable
	if err := os.Chmod(tmpFile, 0755); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to make binary executable: %w", err)
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
// TODO: Remove this function after migration to IsFixPanicAgentInstalled
func (m *Manager) IsInstalled() bool {
	fmt.Println("WARNING: IsInstalled() is deprecated, use IsFixPanicAgentInstalled() instead")
	return m.IsFixPanicAgentInstalled()
}

// GetVersion returns the version of the installed connectivity layer (DEPRECATED)
// TODO: Remove this function after migration to GetFixPanicAgentVersion
func (m *Manager) GetVersion() (string, error) {
	fmt.Println("WARNING: GetVersion() is deprecated, use GetFixPanicAgentVersion() instead")
	return m.GetFixPanicAgentVersion()
}

// Remove removes the connectivity layer binary (DEPRECATED)
// TODO: Remove this function after migration to RemoveFixPanicAgent
func (m *Manager) Remove() error {
	fmt.Println("WARNING: Remove() is deprecated, use RemoveFixPanicAgent() instead")
	return m.RemoveFixPanicAgent()
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

// DownloadFixPanicAgent downloads the FixPanic Agent binary from GitHub Releases
func (m *Manager) DownloadFixPanicAgent(version string) error {
	downloadURL, err := platform.GetFixPanicAgentDownloadURL(version)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	binaryPath := m.platform.GetFixPanicAgentBinaryPath()

	logger.Loading("Downloading from %s...", downloadURL)

	// Create temporary file
	tmpFile := binaryPath + ".tmp"

	resp, err := m.client.Get(downloadURL)
	if err != nil {
		logger.LoadingFailed("Failed to download")
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.LoadingFailed("HTTP %d", resp.StatusCode)
		return fmt.Errorf("failed to download binary: HTTP %d", resp.StatusCode)
	}

	logger.LoadingDone("Download started")

	// Create the file
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to save binary: %w", err)
	}

	// Make the binary executable
	if err := os.Chmod(tmpFile, 0755); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Move to final location
	if err := os.Rename(tmpFile, binaryPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to move binary to final location: %w", err)
	}

	logger.Success("FixPanic Agent downloaded to %s", binaryPath)
	return nil
}

// IsFixPanicAgentInstalled checks if the FixPanic Agent is installed
func (m *Manager) IsFixPanicAgentInstalled() bool {
	binaryPath := m.platform.GetFixPanicAgentBinaryPath()
	_, err := os.Stat(binaryPath)
	return err == nil
}

// GetFixPanicAgentVersion returns the version of the installed FixPanic Agent
func (m *Manager) GetFixPanicAgentVersion() (string, error) {
	binaryPath := m.platform.GetFixPanicAgentBinaryPath()

	if !m.IsFixPanicAgentInstalled() {
		return "", fmt.Errorf("FixPanic Agent not installed")
	}

	// Execute with --version flag
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// UpdateFixPanicAgent updates the FixPanic Agent to the specified version
func (m *Manager) UpdateFixPanicAgent(version string) error {
	fmt.Printf("Updating FixPanic Agent to version %s...\n", version)

	// Remove old version
	if err := m.RemoveFixPanicAgent(); err != nil {
		return fmt.Errorf("failed to remove old version: %w", err)
	}

	// Download new version
	if err := m.DownloadFixPanicAgent(version); err != nil {
		return fmt.Errorf("failed to download new version: %w", err)
	}

	fmt.Printf("FixPanic Agent updated successfully\n")
	return nil
}

// RemoveFixPanicAgent removes the FixPanic Agent binary
func (m *Manager) RemoveFixPanicAgent() error {
	binaryPath := m.platform.GetFixPanicAgentBinaryPath()

	if err := os.Remove(binaryPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already removed
		}
		return fmt.Errorf("failed to remove binary: %w", err)
	}

	return nil
}

// Update updates the connectivity layer to the specified version (DEPRECATED)
// TODO: Remove this function after migration to UpdateFixPanicAgent
func (m *Manager) Update(version string) error {
	fmt.Println("WARNING: Update() is deprecated, use UpdateFixPanicAgent() instead")
	return m.UpdateFixPanicAgent(version)
}

// AgentRelease represents a GitHub release for the agent binary
type AgentRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	PublishedAt string `json:"published_at"`
}

// GetLatestAgentVersion fetches the latest agent version from GitHub releases
func (m *Manager) GetLatestAgentVersion() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	url := "https://api.github.com/repos/fixpanic/fixpanic-connectivity-layer-release/releases/latest"

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API request failed: %d", resp.StatusCode)
	}

	var release AgentRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}

	return release.TagName, nil
}

// IsAgentUpdateAvailable checks if a newer version of the agent is available
func (m *Manager) IsAgentUpdateAvailable() (bool, string, error) {
	if !m.IsFixPanicAgentInstalled() {
		return true, "", nil // Need to install
	}

	currentVersion, err := m.GetFixPanicAgentVersion()
	if err != nil {
		return true, "", fmt.Errorf("failed to get current version: %w", err)
	}

	latestVersion, err := m.GetLatestAgentVersion()
	if err != nil {
		return false, "", fmt.Errorf("failed to get latest version: %w", err)
	}

	// Parse version strings to compare them
	// For simplicity, we'll do string comparison since they follow semantic versioning
	currentClean := strings.TrimSpace(currentVersion)
	latestClean := strings.TrimSpace(latestVersion)

	// Extract version from output like "fixpanic-connectivity-layer v1.0.0 - ..."
	if strings.Contains(currentClean, " v") {
		parts := strings.Split(currentClean, " v")
		if len(parts) > 1 {
			versionPart := strings.Split(parts[1], " ")[0]
			currentClean = "v" + versionPart
		}
	}

	return currentClean != latestClean, latestClean, nil
}

// EnsureLatestAgent checks and updates the agent binary if needed
func (m *Manager) EnsureLatestAgent() error {
	logger.Progress("Checking for agent binary updates")

	updateAvailable, latestVersion, err := m.IsAgentUpdateAvailable()
	if err != nil {
		logger.Warning("Failed to check for updates: %v", err)
		// Continue with existing binary if update check fails
		return nil
	}

	if !updateAvailable {
		if m.IsFixPanicAgentInstalled() {
			logger.List("Agent binary is up to date")
		}
		return nil
	}

	// Update or install the agent
	if m.IsFixPanicAgentInstalled() {
		currentVersion, _ := m.GetFixPanicAgentVersion()
		logger.Info("Agent update available: %s → %s", currentVersion, latestVersion)
		logger.Progress("Downloading latest agent binary")
	} else {
		logger.Progress("Installing agent binary")
	}

	if err := m.DownloadFixPanicAgent("latest"); err != nil {
		return fmt.Errorf("failed to download latest agent: %w", err)
	}

	// Verify the update
	newVersion, err := m.GetFixPanicAgentVersion()
	if err != nil {
		logger.Warning("Failed to verify new version: %v", err)
	} else {
		logger.Success("Agent binary updated to: %s", newVersion)
	}

	return nil
}
