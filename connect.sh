#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# OpsSquad Free Tool — Temporary Agent Bootstrap
# ============================================================
#
# Usage (called automatically by the free-tool session flow):
#   curl -sSL https://raw.githubusercontent.com/fixpanic/opssquad-cli-tool/main/connect.sh | bash -s -- \
#     --node-id <uuid> --token <token> [--ttl 30m] [--server <host:port>]
#
# This script:
#   1. Validates required --node-id and --token arguments
#   2. Detects OS and architecture
#   3. Downloads the correct agent binary from GitHub Releases
#   4. Creates a minimal YAML config with temporary mode settings
#   5. Starts the agent in the foreground (Ctrl+C to stop)
#   6. Cleans up temp files on exit
# ============================================================

VERSION="1.0.0"
BINARY_NAME="opssquad-connectivity-layer"
GITHUB_REPO="fixpanic/opssquad-connectivity-layer-release"

# ---- Colors ----
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

# ---- Defaults ----
NODE_ID=""
TOKEN=""
TTL=""
SERVER_URL="socket.opssquad.ai:9000"

# ---- Parse arguments ----
while [[ $# -gt 0 ]]; do
    case $1 in
        --node-id)
            [[ -z "${2:-}" ]] && echo -e "${RED}Error: --node-id requires a value${NC}" >&2 && exit 1
            NODE_ID="$2"; shift 2 ;;
        --token)
            [[ -z "${2:-}" ]] && echo -e "${RED}Error: --token requires a value${NC}" >&2 && exit 1
            TOKEN="$2"; shift 2 ;;
        --ttl)
            [[ -z "${2:-}" ]] && echo -e "${RED}Error: --ttl requires a value${NC}" >&2 && exit 1
            TTL="$2"; shift 2 ;;
        --server)
            [[ -z "${2:-}" ]] && echo -e "${RED}Error: --server requires a value${NC}" >&2 && exit 1
            SERVER_URL="$2"; shift 2 ;;
        --temporary) shift 1 ;;
        --help|-h)
            echo "OpsSquad Free Tool — Temporary Agent v${VERSION}"
            echo ""
            echo "Usage:"
            echo "  curl -sSL https://raw.githubusercontent.com/fixpanic/opssquad-cli-tool/main/connect.sh | bash -s -- \\"
            echo "    --node-id <uuid> --token <token> [--server <host:port>]"
            echo ""
            echo "Options:"
            echo "  --node-id <uuid>      Node ID (required, provided by free tool session)"
            echo "  --token <token>        Auth token (required, provided by free tool session)"
            echo "  --ttl <duration>       Session TTL (display only, enforced server-side)"
            echo "  --server <host:port>   Socket server (default: socket.opssquad.ai:9000)"
            echo "  --help, -h             Show this help message"
            exit 0 ;;
        *)
            echo -e "${RED}Error: Unknown option: $1${NC}" >&2
            echo "Run with --help for usage information." >&2
            exit 1 ;;
    esac
done

# ---- Validate required arguments ----
if [[ -z "$NODE_ID" ]]; then
    echo -e "${RED}Error: --node-id is required${NC}" >&2
    echo "Usage: curl -sSL <url> | bash -s -- --node-id <uuid> --token <token>" >&2
    exit 1
fi

if [[ -z "$TOKEN" ]]; then
    echo -e "${RED}Error: --token is required${NC}" >&2
    echo "Usage: curl -sSL <url> | bash -s -- --node-id <uuid> --token <token>" >&2
    exit 1
fi

# ---- Header ----
echo ""
echo -e "${BOLD}${BLUE}OpsSquad Free Tool Agent v${VERSION}${NC}"
echo -e "----------------------------------------------"
echo -e "  Node ID:  ${NODE_ID:0:8}...${NODE_ID: -4}"
echo -e "  Token:    ${TOKEN:0:8}...${TOKEN: -4}"
echo -e "  Server:   ${SERVER_URL}"
[[ -n "$TTL" ]] && echo -e "  TTL:      ${TTL}"
echo -e "  Mode:     temporary (auto-cleanup on disconnect)"
echo -e "----------------------------------------------"
echo ""

# ---- Detect OS ----
OS_RAW=$(uname -s)
case "$OS_RAW" in
    Linux)  OS="linux" ;;
    Darwin) OS="darwin" ;;
    *)
        echo -e "${RED}Error: Unsupported OS: ${OS_RAW}${NC}" >&2
        echo "Supported: Linux, macOS" >&2
        exit 1 ;;
esac

# ---- Detect architecture ----
ARCH_RAW=$(uname -m)
case "$ARCH_RAW" in
    x86_64)         ARCH="amd64" ;;
    aarch64|arm64)  ARCH="arm64" ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: ${ARCH_RAW}${NC}" >&2
        echo "Supported: x86_64 (amd64), aarch64/arm64" >&2
        exit 1 ;;
esac

echo -e "${GREEN}Detected: ${OS}/${ARCH}${NC}"

# ---- Temp directory with cleanup ----
TEMP_DIR=$(mktemp -d -t opssquad-agent-XXXXXX)

cleanup() {
    local exit_code=$?
    echo ""
    echo -e "${YELLOW}Cleaning up temporary files...${NC}"
    rm -rf "$TEMP_DIR"
    if [[ $exit_code -eq 0 ]]; then
        echo -e "${GREEN}Session ended cleanly.${NC}"
    else
        echo -e "${YELLOW}Session ended (exit code: ${exit_code}).${NC}"
    fi
}
trap cleanup EXIT

# ---- Download binary from GitHub Releases ----
ARTIFACT="${BINARY_NAME}-${OS}-${ARCH}"
BINARY_URL="https://github.com/${GITHUB_REPO}/releases/latest/download/${ARTIFACT}"
BINARY_PATH="${TEMP_DIR}/${BINARY_NAME}"

echo -e "${BLUE}Downloading agent binary...${NC}"

HTTP_CODE=$(curl -sSL -w "%{http_code}" -o "${BINARY_PATH}" "$BINARY_URL" 2>/dev/null || true)

if [[ ! -f "$BINARY_PATH" ]] || [[ ! -s "$BINARY_PATH" ]]; then
    echo -e "${RED}Error: Failed to download agent binary${NC}" >&2
    echo -e "  URL: ${BINARY_URL}" >&2
    echo -e "  HTTP: ${HTTP_CODE:-unknown}" >&2
    echo "Check: https://github.com/${GITHUB_REPO}/releases" >&2
    exit 1
fi

if [[ "${HTTP_CODE}" != "200" ]]; then
    echo -e "${RED}Error: Download returned HTTP ${HTTP_CODE}${NC}" >&2
    exit 1
fi

chmod +x "${BINARY_PATH}"

# Remove macOS quarantine attribute
[[ "$OS" == "darwin" ]] && xattr -d com.apple.quarantine "${BINARY_PATH}" 2>/dev/null || true

echo -e "${GREEN}Binary downloaded${NC}"

# ---- Create config ----
CONFIG_FILE="${TEMP_DIR}/config.yaml"
cat > "$CONFIG_FILE" << YAML
# OpsSquad Temporary Agent Configuration
# Auto-generated by connect.sh v${VERSION} — deleted on exit.

app:
  node_id: "${NODE_ID}"
  token: "${TOKEN}"
  server_url: "${SERVER_URL}"
  heartbeat_interval: "15s"
  reconnect_timeout: "5s"
  max_reconnect_attempts: 0
  temporary: true
  tls_enabled: true

logging:
  level: "info"
  format: "text"
  output: "stdout"

req_handler:
  max_concurrent_connections: 10
  connection_timeout: "30s"
  default_tool_timeout: 30
  max_response_size: 524288
  tls_enabled: true
YAML

echo -e "${GREEN}Configuration created${NC}"

# ---- Start ----
echo ""
echo -e "${BOLD}${GREEN}Connecting to OpsSquad...${NC}"
echo -e "${YELLOW}Press Ctrl+C to disconnect and clean up${NC}"
echo ""

exec "${BINARY_PATH}" --config "$CONFIG_FILE" --temporary
