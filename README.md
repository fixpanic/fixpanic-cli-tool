# OpsSquad CLI

> **One command to deploy AI-powered server nodes anywhere**

The OpsSquad CLI is a professional deployment tool for installing and managing AI-powered autonomous nodes on your servers. Deploy intelligent monitoring and troubleshooting capabilities in minutes.

---

## 🚀 Quick Install

```bash
curl -fsSL https://install.opssquad.com/install.sh | bash
```

**That's it!** The script automatically:

- ✅ Detects your platform (Linux/macOS/Windows)
- ✅ Downloads the latest version
- ✅ Installs to the correct location
- ✅ Adds to your PATH

---

## 📦 Get Started

### 1. Install a Node

```bash
opssquad node install \
  --node-id="your-node-id" \
  --token="your-token"
```

### 2. Check Status

```bash
opssquad node status
```

### 3. View Logs

```bash
opssquad node logs --follow
```

---

## 💡 Key Features

| Feature                     | Description                                                |
| --------------------------- | ---------------------------------------------------------- |
| **🤖 AI-Powered**           | Autonomous nodes that understand natural language requests |
| **🔒 Security First**       | Sandboxed execution with command whitelisting              |
| **📊 Real-time Monitoring** | System metrics, logs, and health monitoring                |
| **🌐 Cross-Platform**       | Linux, macOS, Windows support                              |
| **⚡ Zero Dependencies**    | Single binary with no external requirements                |
| **🔧 Easy Management**      | Simple CLI for all node operations                         |

---

## 📋 Commands Reference

### Node Management

```bash
# Install node
opssquad node install --node-id=<id> --token=<token>

# Check status
opssquad node status

# Start/stop node
opssquad node start
opssquad node stop

# View logs
opssquad node logs [--follow] [--lines=100]

# Validate installation
opssquad node validate

# Uninstall
opssquad node uninstall [--force]
```

### Get Help

```bash
opssquad --help
opssquad node --help
opssquad node install --help
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

| Platform    | Architecture                 | Status          |
| ----------- | ---------------------------- | --------------- |
| **Linux**   | amd64, arm64, 386, arm       | ✅ Full Support |
| **macOS**   | amd64 (Intel), arm64 (M1/M2) | ✅ Full Support |
| **Windows** | amd64                        | ✅ Full Support |

**Requirements:**

- Network access to `socket.opssquad.com:9000`
- 50MB disk space
- Linux: systemd (optional, for service management)

---

## 🔧 Manual Installation

### Download Binary

```bash
# Linux (amd64)
curl -LO https://github.com/fixpanic/opssquad-cli-tool/releases/latest/download/opssquad-linux-amd64.tar.gz
tar -xzf opssquad-linux-amd64.tar.gz
sudo mv opssquad /usr/local/bin/

# macOS (arm64)
curl -LO https://github.com/fixpanic/opssquad-cli-tool/releases/latest/download/opssquad-darwin-arm64.tar.gz
tar -xzf opssquad-darwin-arm64.tar.gz
sudo mv opssquad /usr/local/bin/

# Verify installation
opssquad --version
```

### Build from Source

```bash
git clone https://github.com/fixpanic/opssquad-cli-tool.git
cd opssquad-cli-tool
go build -o opssquad
sudo mv opssquad /usr/local/bin/
```

---

## 🔍 Configuration

The node creates configuration files automatically:

### System Installation (root)

```
/usr/local/lib/opssquad/opssquad-connectivity-layer
/etc/opssquad/node.yaml
/var/log/opssquad/node.log
```

### User Installation (non-root)

```
~/.local/lib/opssquad/opssquad-connectivity-layer
~/.config/opssquad/node.yaml
~/.local/log/opssquad/node.log
```

### Configuration Format

```yaml
app:
  node_id: "your-node-id"
  token: "your-token"
logging:
  level: "info"
  file: "/var/log/opssquad/node.log"
```

---

## 🆘 Troubleshooting

### Common Issues

**Node won't start?**

```bash
opssquad node validate
opssquad node logs
```

**Connection problems?**

```bash
# Test network connectivity
curl -I socket.opssquad.com:9000
# Check firewall/proxy settings
```

**Permission errors?**

```bash
# Use sudo for system-wide install
sudo opssquad node install --node-id=<id> --token=<token>

# Or install in user directory (default)
opssquad node install --node-id=<id> --token=<token>
```

---

## 📞 Support

- 📧 **Email**: [support@opssquad.com](mailto:support@opssquad.com)
- 📖 **Docs**: [docs.opssquad.com](https://docs.opssquad.com)
- 🐛 **Issues**: [GitHub Issues](https://github.com/fixpanic/opssquad-cli-tool/issues)
- 💬 **Community**: [Discord](https://discord.gg/opssquad)

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

<div align="center">

**[Get Started Now](https://install.opssquad.com) • [Documentation](https://docs.opssquad.com) • [GitHub](https://github.com/fixpanic/opssquad-cli-tool)**

</div>
