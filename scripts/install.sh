#!/bin/bash
# Fixpanic CLI Installation Script
# This script downloads and installs the Fixpanic CLI tool

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GITHUB_REPO="fixpanic/fixpanic-cli"
BINARY_NAME="fixpanic"
INSTALL_DIR="/usr/local/bin"
USER_INSTALL_DIR="$HOME/.local/bin"
VERSION="${VERSION:-latest}"

# Functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"
    
    case "$OS" in
        Linux*)     PLATFORM="linux" ;;
        Darwin*)    PLATFORM="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) PLATFORM="windows" ;;
        *)          PLATFORM="unknown" ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv7*) ARCH="arm" ;;
        i386|i686) ARCH="386" ;;
        *) ARCH="unknown" ;;
    esac
    
    if [ "$PLATFORM" = "unknown" ] || [ "$ARCH" = "unknown" ]; then
        print_error "Unsupported platform: $OS $ARCH"
        exit 1
    fi
    
    ARTIFACT_NAME="${BINARY_NAME}-${PLATFORM}-${ARCH}"
    if [ "$PLATFORM" = "windows" ]; then
        ARTIFACT_NAME="${ARTIFACT_NAME}.exe"
    fi
}

check_dependencies() {
    print_info "Checking dependencies..."
    
    # Check for curl or wget
    if command -v curl >/dev/null 2>&1; then
        DOWNLOAD_CMD="curl -fsSL"
    elif command -v wget >/dev/null 2>&1; then
        DOWNLOAD_CMD="wget -qO-"
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
    
    print_success "Dependencies check passed"
}

get_latest_version() {
    if [ "$VERSION" = "latest" ]; then
        print_info "Fetching latest version..."
        
        LATEST_URL="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
        
        if command -v curl >/dev/null 2>&1; then
            VERSION=$(curl -s "$LATEST_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        elif command -v wget >/dev/null 2>&1; then
            VERSION=$(wget -qO- "$LATEST_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        fi
        
        if [ -z "$VERSION" ]; then
            print_error "Failed to fetch latest version"
            exit 1
        fi
    fi
    
    print_info "Version: $VERSION"
}

download_binary() {
    print_info "Downloading Fixpanic CLI..."
    
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${ARTIFACT_NAME}"
    TEMP_FILE="/tmp/${BINARY_NAME}-$$"
    
    print_info "Download URL: $DOWNLOAD_URL"
    
    if $DOWNLOAD_CMD "$DOWNLOAD_URL" > "$TEMP_FILE"; then
        print_success "Download completed"
    else
        print_error "Download failed"
        rm -f "$TEMP_FILE"
        exit 1
    fi
    
    # Make the binary executable
    chmod +x "$TEMP_FILE"
    
    # Verify the binary
    if [ -x "$TEMP_FILE" ]; then
        print_success "Binary verification passed"
    else
        print_error "Binary verification failed"
        rm -f "$TEMP_FILE"
        exit 1
    fi
    
    BINARY_PATH="$TEMP_FILE"
}

install_binary() {
    print_info "Installing Fixpanic CLI..."
    
    # Determine installation directory
    if [ -w "$INSTALL_DIR" ]; then
        TARGET_DIR="$INSTALL_DIR"
    else
        TARGET_DIR="$USER_INSTALL_DIR"
        
        # Create user bin directory if it doesn't exist
        if [ ! -d "$USER_INSTALL_DIR" ]; then
            mkdir -p "$USER_INSTALL_DIR"
            print_info "Created directory: $USER_INSTALL_DIR"
        fi
        
        # Add to PATH if not already there
        if ! echo "$PATH" | grep -q "$USER_INSTALL_DIR"; then
            print_warning "Please add $USER_INSTALL_DIR to your PATH"
            print_info "Add this to your shell profile (.bashrc, .zshrc, etc.):"
            echo "export PATH=\"$USER_INSTALL_DIR:\$PATH\""
        fi
    fi
    
    TARGET_PATH="$TARGET_DIR/$BINARY_NAME"
    
    # Remove existing binary if it exists
    if [ -f "$TARGET_PATH" ]; then
        print_info "Removing existing binary..."
        rm -f "$TARGET_PATH"
    fi
    
    # Move the binary to target location
    if mv "$BINARY_PATH" "$TARGET_PATH"; then
        print_success "Installation completed"
    else
        print_error "Installation failed"
        rm -f "$BINARY_PATH"
        exit 1
    fi
    
    # Verify installation
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        print_success "Fixpanic CLI installed successfully"
        print_info "Run '$BINARY_NAME --help' to get started"
    else
        print_warning "Installation completed but binary not found in PATH"
        print_info "You may need to restart your shell or add $TARGET_DIR to your PATH"
    fi
}

cleanup() {
    rm -f "/tmp/${BINARY_NAME}-$$"
}

# Main installation process
main() {
    print_info "Fixpanic CLI Installation Script"
    print_info "================================"
    
    # Detect platform
    detect_platform
    print_info "Platform: $PLATFORM"
    print_info "Architecture: $ARCH"
    
    # Check dependencies
    check_dependencies
    
    # Get version
    get_latest_version
    
    # Download binary
    download_binary
    
    # Install binary
    install_binary
    
    # Cleanup
    cleanup
    
    print_success "Installation completed successfully!"
    print_info "Next steps:"
    echo "  1. Run 'fixpanic agent install --agent-id=<your-agent-id> --api-key=<your-api-key>' to install an agent"
    echo "  2. Run 'fixpanic agent status' to check agent status"
    echo "  3. Run 'fixpanic --help' for more commands"
}

# Set up trap for cleanup
trap cleanup EXIT

# Run main function
main "$@"