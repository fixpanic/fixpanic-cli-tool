package connectivity

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

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

// IsInstalled checks if the connectivity layer is installed
func (m *Manager) IsInstalled() bool {
	binaryPath := m.platform.GetBinaryPath()
	_, err := os.Stat(binaryPath)
	return err == nil
}

// GetVersion returns the version of the installed connectivity layer
func (m *Manager) GetVersion() (string, error) {
	binaryPath := m.platform.GetBinaryPath()

	// Check if binary exists
	if !m.IsInstalled() {
		return "", fmt.Errorf("connectivity layer not installed")
	}

	// Execute with --version flag
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// Remove removes the connectivity layer binary
func (m *Manager) Remove() error {
	binaryPath := m.platform.GetBinaryPath()

	if err := os.Remove(binaryPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already removed
		}
		return fmt.Errorf("failed to remove binary: %w", err)
	}

	return nil
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

// Update updates the connectivity layer to the specified version
func (m *Manager) Update(version string) error {
	fmt.Printf("Updating connectivity layer to version %s...\n", version)

	// Remove old version
	if err := m.Remove(); err != nil {
		return fmt.Errorf("failed to remove old version: %w", err)
	}

	// Download new version
	if err := m.Download(version); err != nil {
		return fmt.Errorf("failed to download new version: %w", err)
	}

	fmt.Printf("Connectivity layer updated successfully\n")
	return nil
}
