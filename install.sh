#!/bin/bash
# ppopcode Installation Script for Linux/Mac
# This script installs ppopcode globally so you can run it from anywhere

set -e

INSTALL_DIR="$HOME/bin"
BINARY_NAME="ppopcode"
PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_info() {
    echo -e "${CYAN}→ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

is_in_path() {
    case ":$PATH:" in
        *":$1:"*) return 0 ;;
        *) return 1 ;;
    esac
}

add_to_path() {
    local dir="$1"
    
    if is_in_path "$dir"; then
        print_info "Already in PATH: $dir"
        return
    fi
    
    print_info "Adding to PATH: $dir"
    
    # Detect shell
    local shell_rc=""
    if [ -n "$BASH_VERSION" ]; then
        shell_rc="$HOME/.bashrc"
    elif [ -n "$ZSH_VERSION" ]; then
        shell_rc="$HOME/.zshrc"
    else
        shell_rc="$HOME/.profile"
    fi
    
    # Add to shell rc
    echo "" >> "$shell_rc"
    echo "# ppopcode" >> "$shell_rc"
    echo "export PATH=\"\$PATH:$dir\"" >> "$shell_rc"
    
    print_success "Added to $shell_rc"
}

install_ppopcode() {
    echo -e "${YELLOW}\n=== Installing ppopcode ===${NC}"
    
    # Check if binary exists
    local binary_path="$PROJECT_ROOT/$BINARY_NAME"
    if [ ! -f "$binary_path" ]; then
        print_error "Binary not found: $binary_path"
        print_info "Please build first: go build -o ppopcode ./cmd/ppopcode"
        exit 1
    fi
    
    # Create install directory
    if [ ! -d "$INSTALL_DIR" ]; then
        print_info "Creating directory: $INSTALL_DIR"
        mkdir -p "$INSTALL_DIR"
        print_success "Created directory"
    fi
    
    # Copy binary
    print_info "Copying $BINARY_NAME to $INSTALL_DIR"
    cp "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    print_success "Copied binary"
    
    # Add to PATH
    add_to_path "$INSTALL_DIR"
    
    echo -e "${GREEN}\n=== Installation Complete! ===${NC}"
    echo ""
    echo -e "${YELLOW}To use ppopcode in your current terminal, run:${NC}"
    echo -e "${CYAN}  export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
    echo ""
    echo -e "${YELLOW}Or simply open a new terminal and type:${NC}"
    echo -e "${CYAN}  ppopcode${NC}"
    echo ""
}

uninstall_ppopcode() {
    echo -e "${YELLOW}\n=== Uninstalling ppopcode ===${NC}"
    
    # Remove binary
    local installed_binary="$INSTALL_DIR/$BINARY_NAME"
    if [ -f "$installed_binary" ]; then
        print_info "Removing: $installed_binary"
        rm -f "$installed_binary"
        print_success "Removed binary"
    else
        print_info "Binary not found: $installed_binary"
    fi
    
    # Remove from PATH if directory is empty
    if [ -d "$INSTALL_DIR" ]; then
        if [ -z "$(ls -A "$INSTALL_DIR")" ]; then
            print_info "Removing empty directory: $INSTALL_DIR"
            rmdir "$INSTALL_DIR"
            print_success "Removed directory"
            print_info "Please manually remove PATH entry from your shell rc file"
        else
            print_info "Directory not empty, keeping: $INSTALL_DIR"
        fi
    fi
    
    echo -e "${GREEN}\n=== Uninstallation Complete! ===${NC}"
}

# Main
case "${1:-}" in
    uninstall)
        uninstall_ppopcode
        ;;
    *)
        install_ppopcode
        ;;
esac
