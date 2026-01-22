#!/bin/bash
# OpsSquad CLI Installation Script
# This script downloads and installs the OpsSquad CLI tool

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GITHUB_REPO="fixpanic/opssquad-cli-tool"
BINARY_NAME="opssquad"
INSTALL_DIR="/usr/local/bin"
USER_INSTALL_DIR="${HOME:-/root}/.local/bin"
VERSION="${VERSION:-latest}"

# Detect if running in a container
IS_CONTAINER=false
if [ -f "/.dockerenv" ] || [ -f "/run/.containerenv" ] || grep -q "docker\|kubepod\|containerd" /proc/1/cgroup 2>/dev/null; then
    IS_CONTAINER=true
fi

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

update_shell_profile() {
    # Update shell profiles to include USER_INSTALL_DIR in PATH
    local path_export="export PATH=\"$USER_INSTALL_DIR:\$PATH\""
    local updated=false

    # Detect which shell profiles exist and update them
    for profile in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.zshrc" "$HOME/.profile"; do
        if [ -f "$profile" ]; then
            # Check if already added
            if ! grep -q "$USER_INSTALL_DIR" "$profile" 2>/dev/null; then
                echo "" >> "$profile"
                echo "# Added by OpsSquad CLI installer" >> "$profile"
                echo "$path_export" >> "$profile"
                print_info "Updated $profile with PATH"
                updated=true
            fi
        fi
    done

    # If no profile exists, create .profile (POSIX standard)
    if [ "$updated" = false ]; then
        local profile="$HOME/.profile"
        echo "# Added by OpsSquad CLI installer" > "$profile"
        echo "$path_export" >> "$profile"
        print_info "Created $profile with PATH"
    fi
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
    print_info "Downloading OpsSquad CLI..."

    # Download the .tar.gz archive for Unix, or .exe for Windows
    if [ "$PLATFORM" = "windows" ]; then
        ARCHIVE_NAME="${ARTIFACT_NAME}"
    else
        ARCHIVE_NAME="${ARTIFACT_NAME}.tar.gz"
    fi

    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"
    TEMP_DIR="/tmp/opssquad-install-$$"
    mkdir -p "$TEMP_DIR"
    ARCHIVE_PATH="$TEMP_DIR/$ARCHIVE_NAME"

    print_info "Download URL: $DOWNLOAD_URL"

    # Download with better error handling
    if $DOWNLOAD_CMD "$DOWNLOAD_URL" > "$ARCHIVE_PATH"; then
        print_success "Download completed"
    else
        print_error "Download failed. Please check:"
        print_error "1. Version $VERSION exists"
        print_error "2. Asset $ARCHIVE_NAME is available"
        print_error "3. Network connectivity"
        rm -rf "$TEMP_DIR"
        exit 1
    fi

    # Verify download
    if [ ! -s "$ARCHIVE_PATH" ]; then
        print_error "Downloaded file is empty"
        rm -rf "$TEMP_DIR"
        exit 1
    fi

    if [ "$PLATFORM" = "windows" ]; then
        BINARY_PATH="$ARCHIVE_PATH"
        chmod +x "$BINARY_PATH"
    else
        # Extract the binary from the tar.gz
        print_info "Extracting archive..."
        if tar -xzf "$ARCHIVE_PATH" -C "$TEMP_DIR"; then
            print_success "Extraction completed"
        else
            print_error "Failed to extract archive"
            rm -rf "$TEMP_DIR"
            exit 1
        fi
        
        BINARY_PATH="$TEMP_DIR/${ARTIFACT_NAME}"
        
        # Check if extracted binary exists
        if [ ! -f "$BINARY_PATH" ]; then
            print_error "Binary not found after extraction: ${ARTIFACT_NAME}"
            print_info "Available files in archive:"
            tar -tzf "$ARCHIVE_PATH" || true
            rm -rf "$TEMP_DIR"
            exit 1
        fi
        
        chmod +x "$BINARY_PATH"
    fi

    # Verify the binary
    if [ -x "$BINARY_PATH" ]; then
        print_success "Binary verification passed"
        
        # Test if binary can run (basic compatibility check)
        if "$BINARY_PATH" --version >/dev/null 2>&1 || [ $? -eq 1 ]; then
            print_success "Binary compatibility verified"
        else
            print_error "Binary appears to be incompatible with this system"
            print_error "This might be due to architecture mismatch"
            rm -rf "$TEMP_DIR"
            exit 1
        fi
    else
        print_error "Binary verification failed - file is not executable"
        rm -rf "$TEMP_DIR"
        exit 1
    fi
}

install_binary() {
    print_info "Installing OpsSquad CLI..."

    # In containers, prefer creating /usr/local/bin if it doesn't exist (running as root)
    if [ "$IS_CONTAINER" = true ] && [ "$(id -u)" = "0" ]; then
        print_info "Detected container environment (running as root)"
        if [ ! -d "$INSTALL_DIR" ]; then
            mkdir -p "$INSTALL_DIR"
            print_info "Created $INSTALL_DIR"
        fi
    fi

    # Determine installation directory
    # Priority: /usr/local/bin (if writable) > /usr/bin (container fallback) > ~/.local/bin
    if [ -d "$INSTALL_DIR" ] && [ -w "$INSTALL_DIR" ]; then
        TARGET_DIR="$INSTALL_DIR"
    elif [ -w "/usr/bin" ]; then
        # Fallback for minimal containers where /usr/local/bin might not exist
        TARGET_DIR="/usr/bin"
        print_info "Using /usr/bin as installation directory"
    else
        TARGET_DIR="$USER_INSTALL_DIR"

        # Create user bin directory if it doesn't exist
        if [ ! -d "$USER_INSTALL_DIR" ]; then
            mkdir -p "$USER_INSTALL_DIR"
            print_info "Created directory: $USER_INSTALL_DIR"
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
        print_success "Installation completed to $TARGET_PATH"
    else
        print_error "Installation failed"
        rm -f "$BINARY_PATH"
        exit 1
    fi

    # For user installs, update PATH in current session and shell profiles
    if [ "$TARGET_DIR" = "$USER_INSTALL_DIR" ]; then
        # Add to PATH for current session
        export PATH="$USER_INSTALL_DIR:$PATH"
        print_info "Added $USER_INSTALL_DIR to current session PATH"

        # Update shell profiles for future sessions
        update_shell_profile
    fi

    # Verify installation
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        print_success "OpsSquad CLI installed successfully"

        # Test basic functionality
        if "$BINARY_NAME" --version >/dev/null 2>&1; then
            print_success "Binary is working correctly"
            VERSION_OUTPUT=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown")
            print_info "Installed version: $VERSION_OUTPUT"
        else
            print_warning "Binary installed but --version command failed"
        fi

        print_info "Run '$BINARY_NAME --help' to get started"
    else
        # Binary not in PATH - try to create a symlink to a directory that IS in PATH
        print_warning "Binary not found in PATH after installation"
        print_info "Attempting to create symlink..."

        SYMLINK_CREATED=false

        # Try common PATH directories for symlink
        for symlink_dir in "/usr/bin" "/bin" "/usr/local/bin"; do
            if [ -d "$symlink_dir" ] && [ -w "$symlink_dir" ] && [ "$symlink_dir" != "$TARGET_DIR" ]; then
                if echo "$PATH" | grep -q "$symlink_dir"; then
                    if ln -sf "$TARGET_PATH" "$symlink_dir/$BINARY_NAME" 2>/dev/null; then
                        print_success "Created symlink: $symlink_dir/$BINARY_NAME -> $TARGET_PATH"
                        SYMLINK_CREATED=true
                        break
                    fi
                fi
            fi
        done

        if [ "$SYMLINK_CREATED" = true ]; then
            # Verify symlink works
            if command -v "$BINARY_NAME" >/dev/null 2>&1; then
                print_success "OpsSquad CLI is now accessible"
                if "$BINARY_NAME" --version >/dev/null 2>&1; then
                    VERSION_OUTPUT=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown")
                    print_info "Installed version: $VERSION_OUTPUT"
                fi
            fi
        else
            # Provide manual instructions
            print_warning "Could not create symlink automatically"
            print_info "The binary was installed to: $TARGET_PATH"

            if [ "$TARGET_DIR" = "$USER_INSTALL_DIR" ]; then
                print_info ""
                print_info "To use immediately, run one of:"
                print_info "  export PATH=\"$USER_INSTALL_DIR:\$PATH\""
                print_info "  source ~/.bashrc  # or ~/.zshrc"
                print_info ""
            else
                print_info "Your PATH: $PATH"
                print_info "This may indicate PATH is misconfigured in this environment"
            fi
            print_info "Or run directly: $TARGET_PATH"
        fi
    fi
}

cleanup() {
    rm -rf "/tmp/opssquad-install-$$"
}

# Main installation process
main() {
    print_info "OpsSquad CLI Installation Script"
    print_info "================================"

    # Log container detection
    if [ "$IS_CONTAINER" = true ]; then
        print_info "Container environment detected"
    fi

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
    echo "  1. Run 'opssquad node install --node-id=<your-node-id> --token=<your-token>' to install a node"
    echo "  2. Run 'opssquad node status' to check node status"
    echo "  3. Run 'opssquad --help' for more commands"
}

# Set up trap for cleanup
trap cleanup EXIT

# Run main function
main "$@"