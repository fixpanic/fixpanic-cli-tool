# Installation

## Quick Install
The fastest way to get started is using our automated installation script. This script detects your operating system and architecture, downloads the correct binary, and installs it to your system path.

```bash
curl -fsSL https://install.fixpanic.com/install.sh | bash
```

**The script automatically:**
- ✅ Detects your platform (Linux/macOS/Windows)
- ✅ Downloads the latest version
- ✅ Installs to the correct location
- ✅ Adds `fixpanic` to your PATH

---

## Manual Installation

If you prefer to install manually, you can download the binaries directly or build from source.

### Linux (amd64)
```bash
curl -LO https://github.com/fixpanic/fixpanic-cli-tool/releases/latest/download/fixpanic-linux-amd64.tar.gz
tar -xzf fixpanic-linux-amd64.tar.gz
sudo mv fixpanic /usr/local/bin/
```

### macOS (arm64 / Apple Silicon)
```bash
curl -LO https://github.com/fixpanic/fixpanic-cli-tool/releases/latest/download/fixpanic-darwin-arm64.tar.gz
tar -xzf fixpanic-darwin-arm64.tar.gz
sudo mv fixpanic /usr/local/bin/
```

### Windows
*Windows support is fully available. Please download the latest release from our GitHub repository and add the binary to your system PATH.*

### Build from Source
**Prerequisites:** Go 1.21 or later.

```bash
git clone https://github.com/fixpanic/fixpanic-cli-tool.git
cd fixpanic-cli-tool
go build -o fixpanic
sudo mv fixpanic /usr/local/bin/
```

---

## Requirements
- **Disk Space**: Approximately 50MB.
- **Network**: Outbound access to `socket.fixpanic.com:9000`.
- **Permissions**: Root access is recommended for system-wide service management, but user-level installation is also supported.
