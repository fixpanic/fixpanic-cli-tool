# Installation

## Quick Install
The fastest way to get started is using our automated installation script. This script detects your operating system and architecture, downloads the correct binary, and installs it to your system path.

```bash
curl -fsSL https://install.opssquad.com/install.sh | bash
```

**The script automatically:**
- ✅ Detects your platform (Linux/macOS/Windows)
- ✅ Downloads the latest version
- ✅ Installs to the correct location
- ✅ Adds `opssquad` to your PATH

---

## Manual Installation

If you prefer to install manually, you can download the binaries directly or build from source.

### Linux (amd64)
```bash
curl -LO https://github.com/opssquad/opssquad-cli-tool/releases/latest/download/opssquad-linux-amd64.tar.gz
tar -xzf opssquad-linux-amd64.tar.gz
sudo mv opssquad /usr/local/bin/
```

### macOS (arm64 / Apple Silicon)
```bash
curl -LO https://github.com/opssquad/opssquad-cli-tool/releases/latest/download/opssquad-darwin-arm64.tar.gz
tar -xzf opssquad-darwin-arm64.tar.gz
sudo mv opssquad /usr/local/bin/
```

### Windows
*Windows support is fully available. Please download the latest release from our GitHub repository and add the binary to your system PATH.*

### Build from Source
**Prerequisites:** Go 1.21 or later.

```bash
git clone https://github.com/opssquad/opssquad-cli-tool.git
cd opssquad-cli-tool
go build -o opssquad
sudo mv opssquad /usr/local/bin/
```

---

## Requirements
- **Disk Space**: Approximately 50MB.
- **Network**: Outbound access to `socket.opssquad.com:9000`.
- **Permissions**: Root access is recommended for system-wide service management, but user-level installation is also supported.
