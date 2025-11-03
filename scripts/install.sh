#!/bin/bash
# tf-safe installation script for Linux and macOS

set -e

# Configuration
GITHUB_REPO="your-org/tf-safe"
BINARY_NAME="tf-safe"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect platform and architecture
detect_platform() {
    local os arch
    
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$os" in
        linux) os="linux" ;;
        darwin) os="darwin" ;;
        *) 
            log_error "Unsupported operating system: $os"
            exit 1
            ;;
    esac
    
    arch=$(uname -m)
    case "$arch" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *)
            log_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
    
    echo "${os}-${arch}"
}

# Get latest release version
get_latest_version() {
    local version
    log_info "Fetching latest release version..."
    
    version=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        log_error "Failed to fetch latest version"
        exit 1
    fi
    
    echo "$version"
}

# Download and verify binary
download_binary() {
    local version="$1"
    local platform="$2"
    local url filename tmp_dir
    
    filename="${BINARY_NAME}-${platform}.tar.gz"
    url="https://github.com/$GITHUB_REPO/releases/download/$version/$filename"
    tmp_dir=$(mktemp -d)
    
    log_info "Downloading $filename..."
    
    if ! curl -fsSL "$url" -o "$tmp_dir/$filename"; then
        log_error "Failed to download $filename"
        rm -rf "$tmp_dir"
        exit 1
    fi
    
    log_info "Extracting binary..."
    cd "$tmp_dir"
    tar -xzf "$filename"
    
    # Find the extracted binary
    local binary_file="${BINARY_NAME}-${platform}"
    if [ ! -f "$binary_file" ]; then
        log_error "Binary file not found in archive"
        rm -rf "$tmp_dir"
        exit 1
    fi
    
    echo "$tmp_dir/$binary_file"
}

# Install binary
install_binary() {
    local binary_path="$1"
    local install_path="$INSTALL_DIR/$BINARY_NAME"
    
    log_info "Installing $BINARY_NAME to $install_path..."
    
    # Check if we need sudo
    if [ ! -w "$INSTALL_DIR" ]; then
        if command -v sudo >/dev/null 2>&1; then
            sudo cp "$binary_path" "$install_path"
            sudo chmod 755 "$install_path"
        else
            log_error "Cannot write to $INSTALL_DIR and sudo is not available"
            log_info "Please run as root or install manually"
            exit 1
        fi
    else
        cp "$binary_path" "$install_path"
        chmod 755 "$install_path"
    fi
}

# Verify installation
verify_installation() {
    log_info "Verifying installation..."
    
    if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
        log_warning "$BINARY_NAME not found in PATH"
        log_info "You may need to add $INSTALL_DIR to your PATH"
        log_info "Add this to your shell profile:"
        log_info "  export PATH=\"$INSTALL_DIR:\$PATH\""
        return 1
    fi
    
    local version
    version=$($BINARY_NAME --version 2>/dev/null || echo "unknown")
    log_success "$BINARY_NAME installed successfully!"
    log_info "Version: $version"
    
    return 0
}

# Cleanup function
cleanup() {
    if [ -n "$tmp_dir" ] && [ -d "$tmp_dir" ]; then
        rm -rf "$tmp_dir"
    fi
}

# Main installation function
main() {
    local version platform binary_path
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    log_info "tf-safe installation script"
    log_info "Repository: https://github.com/$GITHUB_REPO"
    echo
    
    # Check dependencies
    for cmd in curl tar; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            log_error "Required command not found: $cmd"
            exit 1
        fi
    done
    
    # Detect platform
    platform=$(detect_platform)
    log_info "Detected platform: $platform"
    
    # Get version (use provided version or fetch latest)
    if [ -n "$1" ]; then
        version="$1"
        log_info "Using specified version: $version"
    else
        version=$(get_latest_version)
        log_info "Latest version: $version"
    fi
    
    # Download binary
    binary_path=$(download_binary "$version" "$platform")
    
    # Install binary
    install_binary "$binary_path"
    
    # Verify installation
    if verify_installation; then
        echo
        log_success "Installation completed successfully!"
        echo
        log_info "Next steps:"
        log_info "1. Initialize a Terraform project: tf-safe init"
        log_info "2. Configure backends in .tf-safe.yaml"
        log_info "3. Use tf-safe instead of terraform: tf-safe apply"
        log_info "4. View help: tf-safe --help"
    else
        echo
        log_warning "Installation completed but verification failed"
        log_info "Try running: $INSTALL_DIR/$BINARY_NAME --version"
    fi
}

# Handle command line arguments
case "${1:-}" in
    -h|--help)
        echo "tf-safe installation script"
        echo
        echo "Usage: $0 [version]"
        echo
        echo "Arguments:"
        echo "  version    Specific version to install (e.g., v1.0.0)"
        echo "             If not specified, installs the latest version"
        echo
        echo "Options:"
        echo "  -h, --help Show this help message"
        echo
        echo "Examples:"
        echo "  $0                # Install latest version"
        echo "  $0 v1.0.0         # Install specific version"
        echo
        echo "Environment variables:"
        echo "  INSTALL_DIR       Installation directory (default: /usr/local/bin)"
        echo
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac