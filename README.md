# FixPanic CLI

> **One command to deploy AI-powered server agents anywhere**

The FixPanic CLI is a professional deployment tool for installing and managing AI-powered autonomous agents on your servers. Deploy intelligent monitoring and troubleshooting capabilities in minutes.

---

## 🚀 Quick Install

```bash
curl -fsSL https://install.fixpanic.com/install.sh | bash
```

**That's it!** The script automatically:
- ✅ Detects your platform (Linux/macOS/Windows)
- ✅ Downloads the latest version
- ✅ Installs to the correct location
- ✅ Adds to your PATH

---

## 📦 Get Started

### 1. Install an Agent
```bash
fixpanic agent install \
  --agent-id="your-agent-id" \
  --api-key="your-api-key"
```

### 2. Check Status
```bash
fixpanic agent status
```

### 3. View Logs
```bash
fixpanic agent logs --follow
```

---

## 💡 Key Features

| Feature | Description |
|---------|-------------|
| **🤖 AI-Powered** | Autonomous agents that understand natural language requests |
| **🔒 Security First** | Sandboxed execution with command whitelisting |
| **📊 Real-time Monitoring** | System metrics, logs, and health monitoring |
| **🌐 Cross-Platform** | Linux, macOS, Windows support |
| **⚡ Zero Dependencies** | Single binary with no external requirements |
| **🔧 Easy Management** | Simple CLI for all agent operations |

---

## 📋 Commands Reference

### Agent Management
```bash
# Install agent
fixpanic agent install --agent-id=<id> --api-key=<key>

# Check status
fixpanic agent status

# Start/stop agent
fixpanic agent start
fixpanic agent stop

# View logs
fixpanic agent logs [--follow] [--lines=100]

# Validate installation
fixpanic agent validate

# Uninstall
fixpanic agent uninstall [--force]
```

### Get Help
```bash
fixpanic --help
fixpanic agent --help
fixpanic agent install --help
```

---

## 🎯 Use Cases

- **DevOps Teams**: Automated server diagnostics and troubleshooting
- **SRE**: Intelligent incident response and root cause analysis
- **Monitoring**: AI-powered system health analysis
- **Support**: Natural language server investigation
- **Compliance**: Automated security and configuration auditing

---

## 🌍 Platform Support

| Platform | Architecture | Status |
|----------|--------------|--------|
| **Linux** | amd64, arm64, 386, arm | ✅ Full Support |
| **macOS** | amd64 (Intel), arm64 (M1/M2) | ✅ Full Support |
| **Windows** | amd64 | ✅ Full Support |

**Requirements:**
- Network access to `socket.fixpanic.com:9000`
- 50MB disk space
- Linux: systemd (optional, for service management)

---

## 🔧 Manual Installation

### Download Binary
```bash
# Linux (amd64)
curl -LO https://github.com/fixpanic/fixpanic-cli-tool/releases/latest/download/fixpanic-linux-amd64.tar.gz
tar -xzf fixpanic-linux-amd64.tar.gz
sudo mv fixpanic /usr/local/bin/

# macOS (arm64)
curl -LO https://github.com/fixpanic/fixpanic-cli-tool/releases/latest/download/fixpanic-darwin-arm64.tar.gz
tar -xzf fixpanic-darwin-arm64.tar.gz
sudo mv fixpanic /usr/local/bin/

# Verify installation
fixpanic --version
```

### Build from Source
```bash
git clone https://github.com/fixpanic/fixpanic-cli-tool.git
cd fixpanic-cli-tool
go build -o fixpanic
sudo mv fixpanic /usr/local/bin/
```

---

## 🔍 Configuration

The agent creates configuration files automatically:

### System Installation (root)
```
/usr/local/lib/fixpanic/fixpanic-connectivity-layer
/etc/fixpanic/agent.yaml
/var/log/fixpanic/agent.log
```

### User Installation (non-root)
```
~/.local/lib/fixpanic/fixpanic-connectivity-layer
~/.config/fixpanic/agent.yaml
~/.local/log/fixpanic/agent.log
```

### Configuration Format
```yaml
app:
  agent_id: "your-agent-id"
  api_key: "your-api-key"
logging:
  level: "info"
  file: "/var/log/fixpanic/agent.log"
```

---

## 🆘 Troubleshooting

### Common Issues

**Agent won't start?**
```bash
fixpanic agent validate
fixpanic agent logs
```

**Connection problems?**
```bash
# Test network connectivity
curl -I socket.fixpanic.com:9000
# Check firewall/proxy settings
```

**Permission errors?**
```bash
# Use sudo for system-wide install
sudo fixpanic agent install --agent-id=<id> --api-key=<key>

# Or install in user directory (default)
fixpanic agent install --agent-id=<id> --api-key=<key>
```

---

## 📞 Support

- 📧 **Email**: [support@fixpanic.com](mailto:support@fixpanic.com)
- 📖 **Docs**: [docs.fixpanic.com](https://docs.fixpanic.com)
- 🐛 **Issues**: [GitHub Issues](https://github.com/fixpanic/fixpanic-cli-tool/issues)
- 💬 **Community**: [Discord](https://discord.gg/fixpanic)

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

<div align="center">

**[Get Started Now](https://install.fixpanic.com) • [Documentation](https://docs.fixpanic.com) • [GitHub](https://github.com/fixpanic/fixpanic-cli-tool)**

</div>