# Release Readiness Report

## Overview

This report summarizes the findings from the code scan of the `fixpanic-cli-tool` codebase. The goal was to verify if the CLI tool is ready for production release.

## Summary

The codebase is generally well-structured and follows Go best practices. However, there are a few critical and high-priority issues that need to be addressed before a production release, particularly regarding file permissions and security on different platforms.

## Findings

### 1. Critical: Hardcoded Log Path for Non-Root Users

**Location:** `internal/config/config.go`
**Issue:** The `DefaultConfig` function hardcodes the log file path to `/var/log/fixpanic/agent.log`.

```go
Logging: LoggingSection{
    Level: "info",
    File:  "/var/log/fixpanic/agent.log",
},
```

**Impact:** When a non-root user installs the agent, this configuration is saved to their user config file. When the agent attempts to start, it will try to write to `/var/log/fixpanic/agent.log`, which will fail with "permission denied" as non-root users do not have write access to `/var/log`.
**Recommendation:** Update `DefaultConfig` or the installation logic to dynamically set the log path based on the user's privileges (e.g., using `platformInfo.LogDir`).

### 2. High: Insecure Log Path in macOS Launchd Service

**Location:** `internal/process/darwin.go`
**Issue:** The `generatePlistContent` function sets the standard output and error paths to `/tmp`.

```xml
<key>StandardOutPath</key>
<string>/tmp/fixpanic-agent.log</string>
<key>StandardErrorPath</key>
<string>/tmp/fixpanic-agent-error.log</string>
```

**Impact:** Files in `/tmp` are world-writable. This creates a security vulnerability where a malicious user could potentially tamper with the logs or create symlinks to other files, leading to potential denial of service or data corruption.
**Recommendation:** Use the user's log directory (e.g., `~/.local/log/fixpanic/`) for these logs, consistent with the rest of the application.

### 3. Medium: Systemd Service Installation Limitations

**Location:** `internal/service/manager.go`
**Issue:** The service installation logic attempts to write to `/etc/systemd/system/`, which requires root privileges.
**Impact:** Non-root users on Linux cannot install the agent as a systemd service. While the installer handles this gracefully by falling back to manual start, it limits the functionality for non-root users (no auto-start on boot).
**Recommendation:** Consider supporting user-level systemd services (`~/.config/systemd/user/`) for non-root installations.

### 4. Low: Hardcoded Socket Server Address

**Location:** `cmd/root.go`
**Issue:** The default socket server address is hardcoded to `socket.fixpanic.com:8080`.

```go
rootCmd.PersistentFlags().String("socket-server", "socket.fixpanic.com:8080", "Socket server address")
```

**Impact:** Ensure this is the correct production endpoint. Port 8080 is often used for development or alternative HTTP. Production services typically use port 443 (HTTPS) or a dedicated secure port.
**Recommendation:** Verify this endpoint and port are correct for the production environment.

### 5. General Observations

- **No TODOs/FIXMEs:** A scan of the codebase revealed no "TODO", "FIXME", or "XXX" comments, indicating no obvious unfinished work.
- **No Hardcoded Credentials:** No hardcoded passwords, secrets, or tokens were found.
- **Dependencies:** The project uses standard Go modules and dependencies.
- **Build System:** The `Makefile` provides comprehensive build, test, and release targets.

## Conclusion

The CLI tool is **NOT** ready for release until the Critical and High issues are resolved. The hardcoded log path will cause immediate failure for non-root users, and the insecure log path on macOS presents a security risk.

Please review and address these issues before proceeding with the release.
