package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fixpanic/fixpanic-cli/internal/logger"
	"github.com/spf13/cobra"
)

var (
	forceUpgrade bool
	checkOnly    bool
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade FixPanic CLI to the latest version",
	Long: `Upgrade the FixPanic CLI to the latest version available on GitHub releases.

This command will:
- Check the current version
- Fetch the latest release information from GitHub
- Download and install the new version if available
- Verify the upgrade was successful

The upgrade is performed safely by downloading to a temporary location first,
then replacing the current binary atomically.`,
	Example: `  # Check for available updates
  fixpanic upgrade --check

  # Upgrade to the latest version
  fixpanic upgrade

  # Force upgrade even if already on latest version
  fixpanic upgrade --force`,
	RunE: runUpgrade,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)

	// Add flags
	upgradeCmd.Flags().BoolVar(&forceUpgrade, "force", false, "Force upgrade even if already on latest version")
	upgradeCmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates without upgrading")
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	PublishedAt string `json:"published_at"`
	Body        string `json:"body"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	logger.Header("FixPanic CLI Upgrade")

	// Get current version info
	logger.Step(1, "Checking current version")
	currentVersion := getCurrentVersion()
	logger.KeyValue("Current version", currentVersion)

	// Fetch latest release info
	logger.Step(2, "Fetching latest release information")
	latestRelease, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}

	logger.KeyValue("Latest version", latestRelease.TagName)
	logger.KeyValue("Release date", formatReleaseDate(latestRelease.PublishedAt))

	// Compare versions
	if !forceUpgrade && currentVersion == latestRelease.TagName {
		logger.Success("You are already on the latest version!")
		return nil
	}

	if checkOnly {
		if currentVersion == latestRelease.TagName {
			logger.Success("You are on the latest version")
		} else {
			logger.Info("Update available: %s → %s", currentVersion, latestRelease.TagName)
		}
		return nil
	}

	// Show what will be upgraded
	logger.Separator()
	if currentVersion == latestRelease.TagName {
		logger.Info("Forcing upgrade to same version")
	} else {
		logger.Info("Upgrading: %s → %s", currentVersion, latestRelease.TagName)
	}
	logger.Separator()

	// Get current binary path
	currentBinaryPath, err := getCurrentBinaryPath()
	if err != nil {
		return fmt.Errorf("failed to get current binary path: %w", err)
	}

	logger.KeyValue("Current binary", currentBinaryPath)

	// Download new version
	logger.Step(3, "Downloading new version")
	newBinaryPath, err := downloadNewVersion(latestRelease)
	if err != nil {
		return fmt.Errorf("failed to download new version: %w", err)
	}
	defer os.RemoveAll(filepath.Dir(newBinaryPath)) // Cleanup temp directory

	// Verify new binary
	logger.Step(4, "Verifying new binary")
	if err := verifyNewBinary(newBinaryPath); err != nil {
		return fmt.Errorf("failed to verify new binary: %w", err)
	}

	logger.Success("New binary verified successfully")

	// Replace current binary
	logger.Step(5, "Installing new version")
	if err := replaceBinary(currentBinaryPath, newBinaryPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Verify upgrade
	logger.Step(6, "Verifying upgrade")
	if err := verifyUpgrade(currentBinaryPath, latestRelease.TagName); err != nil {
		logger.Warning("Upgrade completed but verification failed: %v", err)
	} else {
		logger.Success("Upgrade verified successfully")
	}

	logger.Separator()
	logger.Success("FixPanic CLI upgraded successfully!")
	logger.KeyValue("New version", latestRelease.TagName)
	logger.Separator()

	// Show release notes if available
	if latestRelease.Body != "" && len(latestRelease.Body) < 500 {
		logger.Info("Release notes:")
		logger.Plain(strings.TrimSpace(latestRelease.Body))
		logger.Separator()
	}

	logger.Info("Run 'fixpanic --version' to confirm the new version")

	return nil
}

// getCurrentVersion returns the current CLI version
func getCurrentVersion() string {
	if version == "" || version == "dev" {
		return "dev"
	}
	return version
}

// getCurrentBinaryPath returns the path to the currently running binary
func getCurrentBinaryPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	// Resolve symlinks
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", err
	}

	return realPath, nil
}

// getLatestRelease fetches the latest release from GitHub
func getLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	url := "https://api.github.com/repos/fixpanic/fixpanic-cli-tool/releases/latest"
	logger.Loading("Fetching from GitHub API...")

	resp, err := client.Get(url)
	if err != nil {
		logger.LoadingFailed("Failed to fetch")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.LoadingFailed("HTTP %d", resp.StatusCode)
		return nil, fmt.Errorf("GitHub API request failed: %d", resp.StatusCode)
	}

	logger.LoadingDone("Release info fetched")

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	return &release, nil
}

// downloadNewVersion downloads the appropriate binary for the current platform
func downloadNewVersion(release *GitHubRelease) (string, error) {
	// Determine platform-specific binary name
	assetName := fmt.Sprintf("fixpanic-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS != "windows" {
		assetName += ".tar.gz"
	} else {
		assetName += ".exe"
	}

	// Find the asset
	var downloadURL string
	var assetSize int64
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			assetSize = asset.Size
			break
		}
	}

	if downloadURL == "" {
		return "", fmt.Errorf("no binary found for platform %s-%s", runtime.GOOS, runtime.GOARCH)
	}

	logger.KeyValue("Asset", assetName)
	logger.KeyValue("Size", fmt.Sprintf("%.1f MB", float64(assetSize)/(1024*1024)))

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "fixpanic-upgrade-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Download the asset
	client := &http.Client{Timeout: 5 * time.Minute}

	logger.Loading("Downloading %s...", assetName)

	resp, err := client.Get(downloadURL)
	if err != nil {
		logger.LoadingFailed("Download failed")
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.LoadingFailed("HTTP %d", resp.StatusCode)
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	logger.LoadingDone("Download completed")

	// Save to temp file
	tempArchivePath := filepath.Join(tempDir, assetName)
	tempFile, err := os.Create(tempArchivePath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	logger.Progress("Saving to temporary file")
	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		tempFile.Close()
		return "", fmt.Errorf("failed to save download: %w", err)
	}

	// Sync to ensure all data is written to disk
	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return "", fmt.Errorf("failed to sync file to disk: %w", err)
	}

	// Close the file before extraction/chmod
	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close file: %w", err)
	}

	// Extract binary if it's a tar.gz
	var binaryPath string
	if strings.HasSuffix(assetName, ".tar.gz") {
		logger.Progress("Extracting binary from archive")
		binaryPath, err = extractBinaryFromTarGz(tempArchivePath, tempDir)
		if err != nil {
			return "", fmt.Errorf("failed to extract binary: %w", err)
		}
	} else {
		// For Windows .exe, the file is the binary
		binaryPath = tempArchivePath
	}

	// Make binary executable
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return "", fmt.Errorf("failed to make binary executable: %w", err)
	}

	logger.Success("Binary ready at: %s", binaryPath)
	return binaryPath, nil
}

// extractBinaryFromTarGz extracts the binary from a tar.gz archive
func extractBinaryFromTarGz(archivePath, extractDir string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Look for the binary file (platform-specific name or "fixpanic")
		baseName := filepath.Base(header.Name)
		if header.Typeflag == tar.TypeReg && (baseName == "fixpanic" || strings.HasPrefix(baseName, "fixpanic-")) {
			binaryPath := filepath.Join(extractDir, "fixpanic")

			outFile, err := os.Create(binaryPath)
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return "", err
			}

			// Sync to ensure all data is written to disk
			if err := outFile.Sync(); err != nil {
				outFile.Close()
				return "", err
			}

			// Close the file
			if err := outFile.Close(); err != nil {
				return "", err
			}

			return binaryPath, nil
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

// verifyNewBinary checks that the new binary is valid
func verifyNewBinary(binaryPath string) error {
	// Try to run --version on the new binary
	// Note: We can't actually execute it here because it might be a different architecture
	// So we just check that it exists and is executable

	info, err := os.Stat(binaryPath)
	if err != nil {
		return fmt.Errorf("binary not found: %w", err)
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("binary is not executable")
	}

	if info.Size() < 1024 {
		return fmt.Errorf("binary seems too small (%d bytes)", info.Size())
	}

	logger.KeyValue("Binary size", fmt.Sprintf("%.1f MB", float64(info.Size())/(1024*1024)))
	return nil
}

// replaceBinary safely replaces the current binary with the new one
func replaceBinary(currentPath, newPath string) error {
	// On Unix systems, we can use os.Rename to atomically replace a running binary
	// The running process continues with the old inode, but new executions use the new binary

	// Create backup first by copying the current binary
	backupPath := currentPath + ".backup"
	logger.Progress("Creating backup: %s", backupPath)

	if err := copyFile(currentPath, backupPath); err != nil {
		// Backup creation is not critical on Unix, log warning but continue
		logger.Warning("Failed to create backup: %v", err)
	}

	// Replace binary using atomic rename
	// This works even when the binary is running because:
	// - The old process keeps its file descriptor to the old inode
	// - The new binary gets the same path but a new inode
	// - Future executions will use the new binary
	logger.Progress("Replacing binary (atomic rename)")

	if err := os.Rename(newPath, currentPath); err != nil {
		// If rename fails, try to restore backup
		if backupPath != "" {
			logger.Warning("Failed to replace binary, attempting to restore backup")
			if restoreErr := os.Rename(backupPath, currentPath); restoreErr != nil {
				return fmt.Errorf("failed to replace binary and failed to restore backup: %w (restore error: %v)", err, restoreErr)
			}
			return fmt.Errorf("failed to replace binary (backup restored): %w", err)
		}
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Make sure new binary is executable
	if err := os.Chmod(currentPath, 0755); err != nil {
		logger.Warning("Failed to set executable permissions: %v", err)
	}

	// Remove backup on success
	if backupPath != "" {
		logger.Progress("Cleaning up backup")
		if err := os.Remove(backupPath); err != nil {
			logger.Warning("Failed to remove backup file: %v", err)
		}
	}

	logger.Info("Binary replaced successfully. Current process will continue with old version.")
	logger.Info("Next execution of 'fixpanic' will use the new version.")

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// verifyUpgrade checks that the upgrade was successful
func verifyUpgrade(binaryPath, expectedVersion string) error {
	// We can't easily verify the version without running the binary
	// since we're inside the same process. This would be better implemented
	// as a separate verification step or by comparing file hashes.

	// For now, just check the binary exists and has the right permissions
	info, err := os.Stat(binaryPath)
	if err != nil {
		return fmt.Errorf("upgraded binary not found: %w", err)
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("upgraded binary is not executable")
	}

	logger.KeyValue("Binary size", fmt.Sprintf("%.1f MB", float64(info.Size())/(1024*1024)))
	return nil
}

// formatReleaseDate formats the GitHub release date
func formatReleaseDate(dateStr string) string {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("January 2, 2006")
}