# Fixpanic CLI

A command-line tool for deploying and managing Fixpanic agents on customer servers. The CLI handles agent installation, configuration, and lifecycle management.

## Overview

The Fixpanic CLI tool simplifies the deployment of Fixpanic agents (connectivity layer) on customer servers. It provides a unified interface for:

- Installing and configuring agents
- Managing agent lifecycle (start, stop, status)
- Testing connectivity to Fixpanic infrastructure
- Validating security rules
- Viewing agent logs

## Installation

### Quick Install (Recommended)

```bash
curl -fsSL https://get.fixpanic.com/install.sh | bash
```

### Manual Installation

1. Download the appropriate binary for your platform from the [releases page](https://github.com/fixpanic/fixpanic-cli/releases)
2. Extract and move to your PATH:
   ```bash
   tar -xzf fixpanic-linux-amd64.tar.gz
   sudo mv fixpanic /usr/local/bin/
   chmod +x /usr/local/bin/fixpanic
   ```

### Build from Source

```bash
git clone https://github.com/fixpanic/fixpanic-cli.git
cd fixpanic-cli
go build -o fixpanic
sudo mv fixpanic /usr/local/bin/
```

## Quick Start

1. **Install an agent:**
   ```bash
   fixpanic agent install --agent-id="your-agent-id" --api-key="your-api-key"
   ```

2. **Check agent status:**
   ```bash
   fixpanic agent status
   ```

3. **Start the agent:**
   ```bash
   fixpanic agent start
   ```

## Commands

### Agent Management

#### `fixpanic agent install`
Install the Fixpanic agent with connectivity layer.

**Flags:**
- `--agent-id` (required): Agent ID from Fixpanic dashboard
- `--api-key` (required): Agent API key from Fixpanic dashboard
- `--socket-server`: Custom socket server address
- `--force`: Force reinstall even if agent exists

**Example:**
```bash
fixpanic agent install --agent-id="agent_123" --api-key="fp_abc123xyz"
```

#### `fixpanic agent status`
Check the status of the installed agent.

**Example:**
```bash
fixpanic agent status
```

#### `fixpanic agent start`
Start the agent service.

**Example:**
```bash
fixpanic agent start
```

#### `fixpanic agent stop`
Stop the agent service.

**Example:**
```bash
fixpanic agent stop
```

#### `fixpanic agent uninstall`
Uninstall the agent completely.

**Flags:**
- `--force`: Force uninstall without confirmation

**Example:**
```bash
fixpanic agent uninstall
```

### Development & Testing

#### `fixpanic agent test-connection`
Test connectivity to Fixpanic infrastructure.

**Example:**
```bash
fixpanic agent test-connection
```

#### `fixpanic agent validate-rules`
Validate security rules configuration.

**Example:**
```bash
fixpanic agent validate-rules
```

#### `fixpanic agent logs`
View agent logs.

**Flags:**
- `--lines, -n`: Number of log lines to show (default: 50)
- `--follow, -f`: Follow logs in real-time

**Examples:**
```bash
fixpanic agent logs
fixpanic agent logs --lines=100
fixpanic agent logs --follow
```

## Configuration

The CLI creates configuration files in the following locations:

### System-wide installation (root):
- **Binary:** `/usr/local/lib/fixpanic/connectivity`
- **Config:** `/etc/fixpanic/agent.yaml`
- **Logs:** `/var/log/fixpanic/agent.log`
- **Service:** `/etc/systemd/system/fixpanic-agent.service`

### User installation (non-root):
- **Binary:** `~/.local/lib/fixpanic/connectivity`
- **Config:** `~/.config/fixpanic/agent.yaml`
- **Logs:** `~/.local/log/fixpanic/agent.log`

### Configuration File Format

```yaml
agent:
  id: "agent_123"
  api_key: "fp_abc123xyz"
  socket_server: "socket.fixpanic.com:8080"

security:
  rules_file: "/etc/fixpanic/security-rules.yaml"

logging:
  level: "info"
  file: "/var/log/fixpanic/agent.log"
```

## Security Rules

Security rules define which commands the agent can execute. The default rules file includes:

### Allowed Commands
- System monitoring: `ls`, `ps`, `top`, `df`, `free`, `uptime`
- Network diagnostics: `netstat`, `ss`, `ping`, `traceroute`
- Container tools: `docker ps`, `docker logs`, `kubectl get`, `kubectl logs`
- Log viewing: `journalctl`, `systemctl status`

### Denied Commands
- Destructive operations: `rm -rf`, `dd`, `mkfs`, `fdisk`
- Privilege escalation: `sudo`, `su`, `passwd`
- Service control: `systemctl restart/stop/start`
- System shutdown: `reboot`, `shutdown`, `poweroff`

## Platform Support

### Supported Operating Systems
- Linux (amd64, arm64, 386, arm)
- macOS (amd64, arm64)
- Windows (amd64)

### Requirements
- Systemd (recommended for service management)
- Network connectivity to Fixpanic infrastructure
- Sufficient permissions for installation

## Troubleshooting

### Connection Issues
If the agent cannot connect to Fixpanic infrastructure:

1. **Test connectivity:**
   ```bash
   fixpanic agent test-connection
   ```

2. **Check network settings:**
   - Verify firewall rules
   - Check proxy settings
   - Test DNS resolution

3. **Review logs:**
   ```bash
   fixpanic agent logs
   ```

### Service Issues
If the systemd service fails to start:

1. **Check service status:**
   ```bash
   sudo systemctl status fixpanic-agent
   ```

2. **Check configuration:**
   ```bash
   sudo journalctl -u fixpanic-agent -n 50
   ```

3. **Verify binary permissions:**
   ```bash
   ls -la /usr/local/lib/fixpanic/connectivity
   ```

### Permission Issues
If you encounter permission errors:

1. **Run with sudo for system-wide installation:**
   ```bash
   sudo fixpanic agent install --agent-id="..." --api-key="..."
   ```

2. **Use user installation:**
   ```bash
   fixpanic agent install --agent-id="..." --api-key="..."
   ```

## Development

### Building
```bash
go build -o fixpanic
```

### Testing
```bash
go test ./...
```

### Cross-compilation
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o fixpanic-linux-amd64

# macOS
GOOS=darwin GOARCH=amd64 go build -o fixpanic-darwin-amd64

# Windows
GOOS=windows GOARCH=amd64 go build -o fixpanic-windows-amd64.exe
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support, please contact:
- Email: support@fixpanic.com
- Documentation: https://docs.fixpanic.com
- Issues: https://github.com/fixpanic/fixpanic-cli/issues