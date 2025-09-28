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
- `cmd/` - Cobra CLI commands using the structure:
  - `root.go` - Base command with global flags and config initialization
  - `agent.go` - Main agent command group
  - `agent_*.go` - Specific agent subcommands (install, start, stop, status, logs, etc.)
  - `upgrade.go` - CLI self-upgrade functionality
- `internal/` - Internal packages:
  - `config/` - YAML configuration management for agent settings
  - `connectivity/` - Manages connection to socket.fixpanic.com
  - `logger/` - Logging utilities
  - `platform/` - Platform detection and paths
  - `process/` - Cross-platform process management with platform-specific implementations
  - `service/` - Service/daemon management (systemd, etc.)

### Key Architectural Patterns
- **Cross-platform support**: Uses build constraints and platform-specific files in `internal/process/` (darwin.go, unix.go, windows.go)
- **CLI framework**: Built with Cobra and Viper for command structure and configuration
- **Process management**: Abstract interface with platform-specific implementations for starting/stopping processes
- **Configuration**: YAML-based config with system-wide (/etc/fixpanic/) and user-specific (~/.fixpanic/) paths
- **Agent lifecycle**: Downloads connectivity layer binary, manages systemd services, handles configuration

### Configuration Paths
- System installation: `/etc/fixpanic/agent.yaml`, `/usr/local/lib/fixpanic/`, `/var/log/fixpanic/`
- User installation: `~/.config/fixpanic/agent.yaml`, `~/.local/lib/fixpanic/`, `~/.local/log/fixpanic/`

### Build System
- Uses Go modules with dependencies managed in go.mod
- Makefile provides comprehensive build targets
- GitHub Actions workflow handles multi-platform releases
- Cross-compilation for linux/amd64, linux/arm64, linux/386, linux/arm, darwin/amd64, darwin/arm64, windows/amd64
- Version information embedded at build time using ldflags

## Dependencies
- **Core**: Uses Cobra for CLI, Viper for configuration, YAML for config files
- **Go version**: 1.21+
- **External tools**: golangci-lint for linting (auto-installed by make lint)

## Agent Operations
The CLI manages a "connectivity layer" binary that connects to socket.fixpanic.com:8080 for AI agent communication. Key operations include:
- Download and install agent binary
- Configure agent with API key and agent ID
- Start/stop agent processes
- Monitor agent status and logs
- Validate installation and connectivity