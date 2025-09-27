# FixPanic Agent Binary Integration Corrections Plan

## Executive Summary

This plan outlines the necessary corrections to align the FixPanic CLI tool with the FixPanic Agent binary distribution requirements specified in the task prompt. The current implementation has critical mismatches in binary naming, download URLs, and platform detection that need to be resolved.

## Critical Issues Identified

### 1. Binary Name Mismatch
- **Current**: `connectivity` / `connectivity.exe`
- **Required**: `fixpanic-agent` / `fixpanic-agent.exe`

### 2. Download URL Mismatch
- **Current**: `https://releases.fixpanic.com/connectivity/latest/connectivity-${os}-${arch}`
- **Required**: `https://github.com/fixpanic/fixpanic-agent/releases/latest/download/fixpanic-agent-${OS}-${ARCH}`

### 3. Platform Detection Gaps
- **Current**: Basic Go runtime detection
- **Required**: Enhanced platform mapping as per task prompt specifications

## Implementation Plan

### Phase 1: Platform Detection Enhancement

**File**: `internal/platform/platform.go`

**Changes Required**:
1. Add new function `GetFixPanicAgentPlatformInfo()`
2. Implement proper OS/architecture mapping
3. Add platform validation

```go
// Add this function to internal/platform/platform.go
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
    
    // Normalize architecture names as per task prompt (x86_64 -> amd64)
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

### Phase 2: Download URL Correction

**File**: `internal/platform/platform.go`

**Changes Required**:
1. Replace `GetConnectivityDownloadURL()` with `GetFixPanicAgentDownloadURL()`
2. Update URL construction to match GitHub Releases pattern

```go
// Replace the existing GetConnectivityDownloadURL function
func GetFixPanicAgentDownloadURL(version string) (string, error) {
    os, arch, err := GetFixPanicAgentPlatformInfo()
    if err != nil {
        return "", fmt.Errorf("failed to get platform info: %w", err)
    }
    
    // Construct URL as per task prompt requirements
    baseURL := "https://github.com/fixpanic/fixpanic-agent/releases"
    
    if version == "latest" {
        return fmt.Sprintf("%s/latest/download/fixpanic-agent-%s-%s", baseURL, os, arch), nil
    }
    
    return fmt.Sprintf("%s/download/%s/fixpanic-agent-%s-%s", baseURL, version, os, arch), nil
}
```

### Phase 3: Binary Name Correction

**File**: `internal/platform/platform.go`

**Changes Required**:
1. Replace `GetConnectivityBinaryName()` with `GetFixPanicAgentBinaryName()`
2. Update binary naming convention

```go
// Replace the existing GetConnectivityBinaryName function
func GetFixPanicAgentBinaryName() string {
    if runtime.GOOS == "windows" {
        return "fixpanic-agent.exe"
    }
    return "fixpanic-agent"
}
```

### Phase 4: Path Management Update

**File**: `internal/platform/platform.go`

**Changes Required**:
1. Add new function `GetFixPanicAgentBinaryPath()`
2. Update path generation for new binary name

```go
// Add this function to internal/platform/platform.go
func (p *PlatformInfo) GetFixPanicAgentBinaryPath() string {
    return fmt.Sprintf("%s/%s", p.LibDir, GetFixPanicAgentBinaryName())
}
```

### Phase 5: Connectivity Manager Enhancement

**File**: `internal/connectivity/manager.go`

**Changes Required**:
1. Add new constructor `NewFixPanicAgentManager()`
2. Add `DownloadFromURL()` method
3. Update binary path references

```go
// Add to internal/connectivity/manager.go
type Manager struct {
    platform   *platform.PlatformInfo
    client     *http.Client
    binaryName string
}

func NewFixPanicAgentManager(platform *platform.PlatformInfo) *Manager {
    return &Manager{
        platform:   platform,
        client:     &http.Client{},
        binaryName: "fixpanic-agent",
    }
}

func (m *Manager) DownloadFixPanicAgent(version string) error {
    downloadURL, err := platform.GetFixPanicAgentDownloadURL(version)
    if err != nil {
        return fmt.Errorf("failed to get download URL: %w", err)
    }
    
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

### Phase 6: Agent Install Command Update

**File**: `cmd/agent_install.go`

**Changes Required**:
1. Update download logic to use new binary
2. Update binary path references
3. Add proper error handling

```go
// Update cmd/agent_install.go
func runAgentInstall(cmd *cobra.Command, args []string) error {
    fmt.Println("Installing FixPanic agent...")

    // Get platform information
    platformInfo, err := platform.GetPlatformInfo()
    if err != nil {
        return fmt.Errorf("failed to get platform info: %w", err)
    }

    // Check if running as root for system-wide installation
    if !platformInfo.IsRoot {
        fmt.Println("Warning: Running as non-root user. Agent will be installed in user directories.")
        fmt.Printf("Binary location: %s\n", platformInfo.LibDir)
        fmt.Printf("Config location: %s\n", platformInfo.ConfigDir)
    }

    // Create necessary directories
    if err := platformInfo.CreateDirectories(); err != nil {
        return fmt.Errorf("failed to create directories: %w", err)
    }

    // Check if already installed
    connectivityManager := connectivity.NewFixPanicAgentManager(platformInfo)
    if connectivityManager.IsInstalled() && !forceInstall {
        return fmt.Errorf("agent is already installed. Use --force to reinstall")
    }

    // Download FixPanic Agent binary (corrected)
    fmt.Println("Downloading FixPanic Agent binary...")
    if err := connectivityManager.DownloadFixPanicAgent("latest"); err != nil {
        return fmt.Errorf("failed to download FixPanic Agent binary: %w", err)
    }

    // Create configuration
    agentConfig := config.DefaultConfig()
    agentConfig.Agent.ID = agentID
    agentConfig.Agent.APIKey = agentAPIKey

    // Validate configuration
    if err := agentConfig.Validate(); err != nil {
        return fmt.Errorf("invalid configuration: %w", err)
    }

    // Save configuration
    configPath := platformInfo.GetConfigPath()
    if err := config.SaveConfig(agentConfig, configPath); err != nil {
        return fmt.Errorf("failed to save configuration: %w", err)
    }

    fmt.Printf("Configuration saved to: %s\n", configPath)

    // Install systemd service if available
    if platform.IsSystemdAvailable() {
        serviceManager := service.NewManager(platformInfo)

        // Remove old service if it exists
        if err := serviceManager.Uninstall(); err != nil {
            fmt.Printf("Warning: failed to remove old service: %v\n", err)
        }

        // Install new service
        if err := serviceManager.Install(); err != nil {
            fmt.Printf("Warning: failed to install systemd service: %v\n", err)
            fmt.Println("You can start the agent manually with: fixpanic agent start")
        } else {
            // Enable and start the service
            if err := serviceManager.Enable(); err != nil {
                fmt.Printf("Warning: failed to enable service: %v\n", err)
            }

            if err := serviceManager.Start(); err != nil {
                fmt.Printf("Warning: failed to start service: %v\n", err)
                fmt.Println("You can start the agent manually with: fixpanic agent start")
            } else {
                fmt.Println("Agent service installed and started successfully")
            }
        }
    } else {
        fmt.Println("Systemd not available. You can start the agent manually with: fixpanic agent start")
    }

    fmt.Println("\n✅ FixPanic agent installed successfully!")
    fmt.Printf("Agent ID: %s\n", agentID)
    fmt.Printf("Binary location: %s\n", platformInfo.GetFixPanicAgentBinaryPath())
    fmt.Printf("Config location: %s\n", configPath)

    if platform.IsSystemdAvailable() {
        fmt.Println("\nThe agent will start automatically on system boot.")
        fmt.Println("You can manage the service with:")
        fmt.Printf("  sudo systemctl status %s\n", platform.GetSystemdServiceName())
        fmt.Printf("  sudo systemctl stop %s\n", platform.GetSystemdServiceName())
        fmt.Printf("  sudo systemctl restart %s\n", platform.GetSystemdServiceName())
    }

    return nil
}
```

### Phase 7: Agent Start Command Update

**File**: `cmd/agent_start.go`

**Changes Required**:
1. Update binary path to use new function
2. Add binary existence validation
3. Update service configuration

```go
// Update cmd/agent_start.go
func runAgentStart(cmd *cobra.Command, args []string) error {
    fmt.Println("Starting FixPanic agent...")

    // Get platform information
    platformInfo, err := platform.GetPlatformInfo()
    if err != nil {
        return fmt.Errorf("failed to get platform info: %w", err)
    }

    // Check if FixPanic Agent binary is installed
    binaryPath := platformInfo.GetFixPanicAgentBinaryPath()
    if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
        return fmt.Errorf("FixPanic Agent binary not found at %s. Run 'fixpanic agent install' first", binaryPath)
    }

    // Try to use systemd service if available
    if platform.IsSystemdAvailable() {
        serviceManager := service.NewManager(platformInfo)

        // Check current status
        status, err := serviceManager.Status()
        if err != nil {
            fmt.Printf("Warning: could not check service status: %v\n", err)
        } else if status == "active" {
            fmt.Println("✅ Agent service is already running")
            return nil
        }

        // Start the service
        if err := serviceManager.Start(); err != nil {
            return fmt.Errorf("failed to start service: %w", err)
        }

        fmt.Println("✅ Agent service started successfully")
        fmt.Printf("Service: %s\n", platform.GetSystemdServiceName())

        // Show how to check status
        fmt.Println("\nYou can check the status with:")
        fmt.Printf("  sudo systemctl status %s\n", platform.GetSystemdServiceName())

        return nil
    }

    // Fallback: run the binary directly (not recommended for production)
    fmt.Println("⚠️  Systemd not available. Starting agent directly...")
    fmt.Println("Note: This is not recommended for production use.")
    fmt.Println("The agent will not restart automatically if it crashes.")

    configPath := platformInfo.GetConfigPath()

    fmt.Printf("Starting: %s --config %s\n", binaryPath, configPath)
    fmt.Println("Press Ctrl+C to stop the agent")

    // Execute the binary directly
    execCmd := exec.Command(binaryPath, "--config", configPath)
    execCmd.Stdout = os.Stdout
    execCmd.Stderr = os.Stderr

    if err := execCmd.Run(); err != nil {
        return fmt.Errorf("failed to start agent: %w", err)
    }

    return nil
}
```

### Phase 8: Service Manager Update

**File**: `internal/service/manager.go`

**Changes Required**:
1. Update service file generation to use correct binary path
2. Update binary references

```go
// Update internal/service/manager.go - generateServiceFile function
func (m *Manager) generateServiceFile() (string, error) {
    binaryPath := m.platform.GetFixPanicAgentBinaryPath() // Use new function
    configPath := m.platform.GetConfigPath()

    tmpl := `[Unit]
Description=FixPanic Agent
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

    // ... rest of function remains the same
}
```

### Phase 9: Enhanced Error Handling

**Files**: Multiple files

**Changes Required**:
1. Add specific error codes
2. Implement retry logic
3. Add detailed error messages

```go
// Add to internal/connectivity/manager.go
type DownloadError struct {
    Code    string
    Message string
    Details string
    URL     string
}

func (e *DownloadError) Error() string {
    return fmt.Sprintf("Download error [%s]: %s (URL: %s)", e.Code, e.Message, e.URL)
}

func (m *Manager) DownloadFixPanicAgent(version string) error {
    downloadURL, err := platform.GetFixPanicAgentDownloadURL(version)
    if err != nil {
        return &DownloadError{
            Code:    "INVALID_PLATFORM",
            Message: "Failed to determine platform information",
            Details: err.Error(),
            URL:     "",
        }
    }
    
    binaryPath := m.platform.GetFixPanicAgentBinaryPath()
    
    // Implement retry logic with exponential backoff
    maxRetries := 3
    for attempt := 1; attempt <= maxRetries; attempt++ {
        if err := m.attemptDownload(downloadURL, binaryPath); err == nil {
            return nil
        }
        
        if attempt < maxRetries {
            backoff := time.Duration(attempt) * time.Second
            log_warning("DOWNLOAD_RETRY", 
                fmt.Sprintf("Download attempt %d failed, retrying in %v...", attempt, backoff))
            time.Sleep(backoff)
        }
    }
    
    return &DownloadError{
        Code:    "MAX_RETRIES_EXCEEDED",
        Message: "Failed to download FixPanic Agent after maximum retries",
        Details: "All download attempts failed",
        URL:     downloadURL,
    }
}
```

### Phase 10: Testing and Validation

**Create test files**:

1. `internal/platform/platform_test.go`
2. `internal/connectivity/manager_test.go`
3. `cmd/agent_install_test.go`

**Test scenarios**:
- Platform detection for all supported platforms
- URL generation for different versions
- Binary download and execution
- Error handling and recovery
- Service installation and management

## Migration Timeline

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| Phase 1-3 | 2 days | Platform detection and URL updates |
| Phase 4-6 | 3 days | Manager and command updates |
| Phase 7-8 | 2 days | Service integration |
| Phase 9 | 1 day | Error handling |
| Phase 10 | 2 days | Testing and validation |
| **Total** | **10 days** | **Complete integration** |

## Risk Mitigation

1. **Backward Compatibility**: Maintain old functions temporarily with deprecation warnings
2. **Rollback Plan**: Keep previous implementation in separate branch
3. **Testing Strategy**: Comprehensive testing on all supported platforms
4. **Documentation**: Update all documentation to reflect changes
5. **Communication**: Notify users of upcoming changes

## Success Criteria

✅ **Binary Download**: Successfully downloads `fixpanic-agent` from GitHub Releases  
✅ **Platform Support**: Correctly detects and maps all supported platforms  
✅ **Binary Execution**: Agent starts with correct binary name and parameters  
✅ **Error Handling**: Comprehensive error messages and recovery mechanisms  
✅ **Service Integration**: Systemd service works with new binary  
✅ **Configuration**: YAML config works with new agent binary  

This comprehensive plan ensures the FixPanic CLI tool correctly integrates with the FixPanic Agent binary distribution system as specified in the task requirements.