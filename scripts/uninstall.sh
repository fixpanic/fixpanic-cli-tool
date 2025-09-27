#!/bin/bash
# Fixpanic CLI Uninstall Script
# This script removes the Fixpanic CLI tool from your system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="fixpanic"
INSTALL_DIR="/usr/local/bin"
USER_INSTALL_DIR="$HOME/.local/bin"

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

find_binary() {
    BINARY_PATH=""
    
    # Check if binary is in PATH
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        BINARY_PATH=$(command -v "$BINARY_NAME")
        print_info "Found $BINARY_NAME at: $BINARY_PATH"
        return 0
    fi
    
    # Check common installation locations
    for dir in "$INSTALL_DIR" "$USER_INSTALL_DIR"; do
        if [ -f "$dir/$BINARY_NAME" ]; then
            BINARY_PATH="$dir/$BINARY_NAME"
            print_info "Found $BINARY_NAME at: $BINARY_PATH"
            return 0
        fi
    done
    
    return 1
}

remove_binary() {
    if [ -z "$BINARY_PATH" ]; then
        print_error "No binary path specified"
        return 1
    fi
    
    print_info "Removing $BINARY_PATH..."
    
    if [ -f "$BINARY_PATH" ]; then
        if rm -f "$BINARY_PATH"; then
            print_success "Successfully removed $BINARY_PATH"
            return 0
        else
            print_error "Failed to remove $BINARY_PATH"
            print_info "You may need to run with sudo or check file permissions"
            return 1
        fi
    else
        print_warning "Binary not found at $BINARY_PATH"
        return 1
    fi
}

cleanup_temp_files() {
    print_info "Cleaning up temporary files..."
    
    # Remove any temporary installation files
    for temp_pattern in "/tmp/fixpanic-*" "/tmp/${BINARY_NAME}-*"; do
        if ls $temp_pattern >/dev/null 2>&1; then
            rm -rf $temp_pattern
            print_info "Removed temporary files: $temp_pattern"
        fi
    done
    
    print_success "Temporary files cleaned up"
}

verify_removal() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        REMAINING_PATH=$(command -v "$BINARY_NAME")
        print_warning "Binary still found in PATH at: $REMAINING_PATH"
        print_info "You may have multiple installations or need to restart your shell"
        return 1
    else
        print_success "Binary successfully removed from PATH"
        return 0
    fi
}

# Main uninstall process
main() {
    print_info "Fixpanic CLI Uninstall Script"
    print_info "============================="
    
    # Check if binary exists
    if ! find_binary; then
        print_warning "Fixpanic CLI not found on this system"
        print_info "It may already be uninstalled or installed in a non-standard location"
        
        # Still try to clean up temp files
        cleanup_temp_files
        exit 0
    fi
    
    # Show current version before removal
    if "$BINARY_PATH" --version >/dev/null 2>&1; then
        VERSION_OUTPUT=$("$BINARY_PATH" --version 2>/dev/null || echo "unknown")
        print_info "Current version: $VERSION_OUTPUT"
    fi
    
    # Confirm removal
    echo
    print_warning "This will remove Fixpanic CLI from your system."
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Uninstall cancelled by user"
        exit 0
    fi
    
    # Remove the binary
    if remove_binary; then
        print_success "Binary removal completed"
    else
        print_error "Binary removal failed"
        exit 1
    fi
    
    # Clean up temporary files
    cleanup_temp_files
    
    # Verify removal
    verify_removal
    
    echo
    print_success "Fixpanic CLI has been successfully uninstalled!"
    print_info "You may want to restart your shell or source your profile to update PATH"
    
    # Provide reinstall instructions
    echo
    print_info "To reinstall Fixpanic CLI, run:"
    echo "  curl -fsSL https://get.fixpanic.com/install.sh | bash"
}

# Run main function
main "$@"