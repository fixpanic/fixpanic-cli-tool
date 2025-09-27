# FixPanic Agent Binary Integration Corrections

## Overview

This document outlines the necessary changes to align the CLI tool with the FixPanic Agent binary distribution requirements as specified in the task prompt.

## Current vs Required Implementation

### Binary Download URL Correction

**Current (Incorrect):**
```go
// internal/platform/platform.go:86
baseURL := "https://releases.fixpanic.com/connectivity"
if version == "latest" {
    return fmt.Sprintf("%s/latest/connectivity-%s-%s", baseURL, os, releaseArch)
}
return fmt.Sprintf("%s/%s/connectivity-%s-%s", baseURL, version, os, releaseArch)
```

**Required (Correct):**
```go
// Should match: https://github.com/your-org/fixpanic-agent/releases/latest/download/fixpanic-agent-${OS}-${ARCH}
baseURL := "https://github.com/fixpanic/fixpanic-agent/releases/latest/download"
binaryName := "fixpanic-agent"
return fmt.Sprintf("%s/%s-%s-%s", baseURL, binaryName, os, releaseArch)
```

### Platform Detection Enhancement

**Current (Basic):**
```go
// Uses Go runtime values directly
os := runtime.GOOS
arch := runtime.GOARCH
```

**Required (Enhanced):**
```go
// Should match task prompt platform mapping
func GetFixPanicAgentPlatform() (string, string, error) {
    // Platform detection as per task prompt
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    if [ "$ARCH" = "x86_64" ]; then ARCH="amd64"; fi
    
    // Map to release naming convention
    platformMap := map[string]string{
        "linux":   "linux",
        "darwin":  "darwin", 
        "windows": "windows",
    }
    
    archMap := map[string]string{
        "amd64": "amd64",
        "arm64": "arm64",
        "386":   "386",
        "arm":   "arm",
    }
    
    goos := runtime.GOOS
    goarch := runtime.GOARCH
    
    // Special handling for x86_64 -> amd64 conversion
    if goarch == "amd64" {
        goarch = "amd64"
    }
    
    platform, ok := platformMap[goos]
    if !ok {
        return "", "", fmt.Errorf("unsupported platform: %s", goos)
    }
    
    arch, ok := archMap[goarch]
    if !ok {
        return "", "", fmt.Errorf("unsupported architecture: %s", goarch)
    }
    
    return platform, arch, nil
}
```

## Corrected Implementation

### 1. Enhanced Platform Detection

```go
// internal/platform/platform.go - Add new function
func GetFixPanicAgentPlatformInfo() (os, arch string, err error) {
    // Platform detection matching task prompt requirements
    goos := runtime.GOOS
    goarch := runtime.GOARCH
    
    // Normalize OS names
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
    
    // Normalize architecture names as per task prompt
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
```

### 2. Corrected Download URL Generation

```go
// internal/platform/platform.go - Replace GetConnectivityDownloadURL
func GetFixPanicAgentDownloadURL(version string) (string, error) {
    os, arch, err := GetFixPanicAgentPlatformInfo()
    if err != nil {
        return "", fmt.Errorf("failed to get platform info: %w", err)
    }
    
    // Construct URL as per task prompt
    baseURL := "https://github.com/fixpanic/fixpanic-agent/releases"
    
    if version == "latest" {
        return fmt.Sprintf("%s/latest/download/fixpanic-agent-%s-%s", baseURL, os, arch), nil
    }
    
    return fmt.Sprintf("%s/download/%s/fixpanic-agent-%s-%s", baseURL, version, os, arch), nil
}
```

### 3. Binary Name Correction

```go
// internal/platform/platform.go - Replace GetConnectivityBinaryName
func GetFixPanicAgentBinaryName() string {
    if runtime.GOOS == "windows" {
        return "fixpanic-agent.exe"
    }
    return "fixpanic-agent"
}
```

### 4. Updated Binary Path

```go
// internal/platform/platform.go - Add new function
func (p *PlatformInfo) GetFixPanicAgentBinaryPath() string {
    return fmt.Sprintf("%s/%s", p.LibDir, GetFixPanicAgentBinaryName())
}
```

## Integration with Existing Code

### Updated Agent Install Command

```go
// cmd/agent_install.go - Update the download section
func runAgentInstall(cmd *cobra.Command, args []string) error {
    // ... existing code ...
    
    // Download FixPanic Agent binary (corrected)
    fmt.Println("Downloading FixPanic Agent binary...")
    
    downloadURL, err := platform.GetFixPanicAgentDownloadURL("latest")
    if err != nil {
        return fmt.Errorf("failed to get download URL: %w", err)
    }
    
    // Create new connectivity manager with corrected binary name
    connectivityManager := connectivity.NewManagerWithBinary(platformInfo, "fixpanic-agent")
    
    if err := connectivityManager.DownloadFromURL(downloadURL); err != nil {
        return fmt.Errorf("failed to download FixPanic Agent binary: %w", err)
    }
    
    // ... rest of installation logic ...
}
```

### Updated Agent Start Command

```go
// cmd/agent_start.go - Update binary execution
func runAgentStart(cmd *cobra.Command, args []string) error {
    // ... existing code ...
    
    // Get corrected binary path
    binaryPath := platformInfo.GetFixPanicAgentBinaryPath()
    configPath := platformInfo.GetConfigPath()
    
    // Verify binary exists
    if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
        return fmt.Errorf("FixPanic Agent binary not found at %s. Run 'fixpanic agent install' first", binaryPath)
    }
    
    // Execute with correct binary name
    execCmd := exec.Command(binaryPath, "--config", configPath)
    execCmd.Stdout = os.Stdout
    execCmd.Stderr = os.Stderr
    
    // ... rest of start logic ...
}
```

## Enhanced Connectivity Manager

```go
// internal/connectivity/manager.go - Add new constructor and method
type Manager struct {
    platform     *platform.PlatformInfo
    client       *http.Client
    binaryName   string
}

func NewManagerWithBinary(platform *platform.PlatformInfo, binaryName string) *Manager {
    return &Manager{
        platform:   platform,
        client:     &http.Client{},
        binaryName: binaryName,
    }
}

func (m *Manager) DownloadFromURL(downloadURL string) error {
    binaryPath := m.platform.GetFixPanicAgentBinaryPath()
    
    fmt.Printf("Downloading FixPanic Agent from %s...\n", downloadURL)
    
    // Create temporary file
    tmpFile := binaryPath + ".tmp"
    
    resp, err := m.client.Get(downloadURL)
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
    
    fmt.Printf("FixPanic Agent downloaded to %s\n", binaryPath)
    return nil
}
```

## Platform Support Matrix

Based on the task prompt requirements:

| Platform | Architecture | Binary Name | Status |
|----------|-------------|-------------|---------|
| Linux    | amd64       | fixpanic-agent-linux-amd64 | ✅ Supported |
| Linux    | arm64       | fixpanic-agent-linux-arm64 | ✅ Supported |
| macOS    | amd64       | fixpanic-agent-darwin-amd64 | ✅ Supported |
| Windows  | amd64       | fixpanic-agent-windows-amd64.exe | ✅ Supported |

## Error Handling Enhancements

```go
// Enhanced error handling for binary operations
func (m *Manager) DownloadFromURL(downloadURL string) error {
    // Pre-download validation
    if err := m.validateDownloadURL(downloadURL); err != nil {
        return fmt.Errorf("invalid download URL: %w", err)
    }
    
    // Download with retry logic
    maxRetries := 3
    for attempt := 1; attempt <= maxRetries; attempt++ {
        if err := m.attemptDownload(downloadURL); err == nil {
            return nil
        }
        
        if attempt < maxRetries {
            log_warning("DOWNLOAD_RETRY", 
                fmt.Sprintf("Download attempt %d failed, retrying...", attempt))
            time.Sleep(time.Duration(attempt) * time.Second)
        }
    }
    
    return fmt.Errorf("failed to download after %d attempts", maxRetries)
}
```

## Configuration Alignment

The configuration should match the task prompt structure:

```yaml
agent:
  id: "your-agent-id"           # Unique agent identifier
  api_key: "your-api-key"       # API authentication key
logging:
  level: "info"                 # Log level (debug, info, warn, error)
  file: "/var/log/fixpanic/agent.log"  # Log file path
```

## Testing and Validation

```bash
# Test platform detection
go run -tags=test ./cmd/test_platform_detection.go

# Test download URL generation
go run -tags=test ./cmd/test_download_urls.go

# Test binary execution
go run -tags=test ./cmd/test_binary_execution.go
```

## Migration Path

1. **Phase 1**: Update platform detection and URL generation
2. **Phase 2**: Update connectivity manager to use new binary name
3. **Phase 3**: Update agent install/start commands
4. **Phase 4**: Add comprehensive testing
5. **Phase 5**: Update documentation

This corrected implementation ensures the CLI tool properly downloads and executes the FixPanic Agent binary according to the task requirements.