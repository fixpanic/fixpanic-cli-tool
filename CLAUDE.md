# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

FixPanic CLI is a Go-based command-line tool for deploying and managing AI-powered autonomous agents on servers. The CLI handles agent installation, configuration, lifecycle management, and provides cross-platform support for Linux, macOS, and Windows.

## Development Commands

### Building
```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Create release packages
make release
```

### Testing
```bash
# Run tests
make test

# Run tests with coverage
make test-coverage
```

### Code Quality
```bash
# Format code
make fmt

# Vet code
make vet

# Lint code (installs golangci-lint if needed)
make lint
```

### Development Workflow
```bash
# Install dependencies
make deps

# Update dependencies
make deps-update

# Clean build artifacts
make clean

# Run the CLI locally
make run

# Install locally to /usr/local/bin
make install
```

## Architecture

### Project Structure
- `main.go` - Entry point that sets version info and executes the root command
- `cmd/` - Cobra CLI commands:
  - `root.go` - Base command with global flags and config initialization
  - `agent.go` - Main agent command group
  - `agent_*.go` - Specific agent subcommands (install, start, stop, status, logs, validate, upgrade, uninstall, connection, restart)
  - `upgrade.go` - CLI self-upgrade functionality
- `internal/` - Internal packages:
  - `config/` - YAML configuration management with validation
  - `connectivity/` - Agent binary download and version management from GitHub Releases
  - `logger/` - Pretty logging utilities for CLI output
  - `platform/` - Platform detection, directory paths, and binary URL generation
  - `process/` - Cross-platform process management with build-constrained implementations
  - `service/` - Systemd service management (install, start, stop, enable, logs)

### Key Architectural Patterns

#### Cross-Platform Support
- **Build constraints**: Platform-specific files in `internal/process/` use `// +build` tags:
  - `darwin.go` - macOS-specific process management
  - `unix.go` - Linux/Unix process management
  - `windows.go` - Windows process management
- **ProcessManager interface**: Abstracts platform differences with common API
- **Platform detection**: Runtime detection of OS/arch with normalized names for binary downloads

#### Installation Modes
The CLI supports two installation modes based on user privileges:

**Root/System Installation** (requires sudo):
- Binary: `/usr/local/lib/fixpanic/fixpanic-connectivity-layer`
- Config: `/etc/fixpanic/agent.yaml`
- Logs: `/var/log/fixpanic/agent.log`
- Service: `/etc/systemd/system/fixpanic-connectivity-layer.service`

**User Installation** (non-root, default):
- Binary: `~/.local/lib/fixpanic/fixpanic-connectivity-layer`
- Config: `~/.config/fixpanic/agent.yaml`
- Logs: `~/.local/log/fixpanic/agent.log`
- No systemd service (manual process management)

#### Agent Binary Management
- **Download source**: GitHub Releases at `fixpanic/fixpanic-connectivity-layer-release`
- **URL pattern**: `https://github.com/fixpanic/fixpanic-connectivity-layer-release/releases/latest/download/fixpanic-connectivity-layer-{os}-{arch}`
- **Version checking**: CLI queries GitHub API for latest release and auto-updates on install
- **Checksum verification**: SHA256 validation available via `VerifyChecksum()`
- **macOS quarantine removal**: Automatically runs `xattr -d com.apple.quarantine` to allow execution

### Build System
- Uses Go modules with dependencies managed in go.mod
- Makefile provides comprehensive build targets
- Cross-compilation for linux/amd64, linux/arm64, linux/386, linux/arm, darwin/amd64, darwin/arm64, windows/amd64
- Version information embedded at build time using ldflags: `-X main.version -X main.commit -X main.date`
- Release packages created as tar.gz for Unix/macOS, raw .exe for Windows

## Dependencies
- **Cobra** (`spf13/cobra`) - CLI framework and command structure
- **Viper** (`spf13/viper`) - Configuration management with environment variable support
- **gopkg.in/yaml.v3** - YAML parsing for agent config files
- **Go version**: 1.21+
- **External tools**: golangci-lint for linting (auto-installed by make lint)

## Agent Operations

The CLI manages the "fixpanic-connectivity-layer" binary (referred to as the "agent"):

### Connection Details
- **Socket server**: `socket.fixpanic.com:8080` (configurable via `--socket-server` flag)
- **Protocol**: TCP connection maintained by the agent binary
- **Configuration**: Agent ID and API key stored in YAML config, passed to agent at startup

### Installation Flow
1. Detect platform and determine installation paths (root vs user)
2. Create necessary directories (lib, config, log)
3. Check for existing installation (skip if found, unless `--force`)
4. Query GitHub API for latest agent version
5. Download agent binary from GitHub Releases
6. Remove macOS quarantine attribute if needed
7. Create YAML config with agent ID and API key
8. Install systemd service (if root and systemd available)
9. Enable and start service

### Command Operations
- `install` - Full installation with auto-update check
- `start/stop/restart` - Process/service lifecycle management (systemd or direct process)
- `status` - Check if agent is running, show PID and service status
- `logs` - Tail agent logs (from journalctl or log file)
- `validate` - Verify installation, config, and connectivity
- `upgrade` - Update agent binary to latest version
- `uninstall` - Remove agent binary, config, logs, and service
- `connection` - Test connection to socket server

## Important Implementation Notes

### Deprecated Functions
Several functions in `internal/platform/platform.go` and `internal/connectivity/manager.go` have deprecation warnings:
- `GetConnectivityBinaryName()` → Use `GetFixPanicAgentBinaryName()`
- `GetConnectivityDownloadURL()` → Use `GetFixPanicAgentDownloadURL()`
- `IsInstalled()` → Use `IsFixPanicAgentInstalled()`
- When refactoring, migrate to the new function names to avoid warnings

### Configuration Management
- Config is validated before saving using `AgentConfig.Validate()`
- Required fields: `agent_id` and `api_key`
- Config paths resolved via `platform.GetPlatformInfo()` based on user privileges
- Environment variables can override config via Viper's `AutomaticEnv()`

### Service Management
- Systemd service runs as the installing user (not always root)
- Service configured with `Restart=always` and `RestartSec=10`
- Logs sent to journalctl on systemd systems
- Non-systemd systems require manual process management