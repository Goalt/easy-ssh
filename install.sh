#!/bin/sh

set -e

# Setup colors
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    BOLD='\033[1m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    BOLD=''
    NC=''
fi

info() {
    printf "${BLUE}info:${NC} %s\n" "$1"
}

success() {
    printf "${GREEN}success:${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}warning:${NC} %s\n" "$1"
}

error() {
    printf "${RED}error:${NC} %s\n" "$1" >&2
    exit 1
}

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m | tr '[:upper:]' '[:lower:]')"

case "$ARCH" in
    x86_64|amd64|x86-64|x64)
        ARCH="amd64"
        ;;
    aarch64|arm64|armv8*)
        ARCH="arm64"
        ;;
    i386|i486|i586|i686)
        ARCH="386"
        ;;
    armv7*|armv6*|arm)
        ARCH="arm"
        ;;
    *)
        error "Unsupported architecture: $ARCH"
        ;;
esac

case "$OS" in
    linux*)
        OS="linux"
        ;;
    darwin*)
        OS="darwin"
        ;;
    *)
        error "Unsupported operating system: $OS"
        ;;
esac

REPO="Goalt/easy-ssh"
BINARY_NAME="easy-ssh"

TAG="${TAG:-${VERSION}}"

if [ -n "$TAG" ]; then
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/easy-ssh-${OS}-${ARCH}"
else
    DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/easy-ssh-${OS}-${ARCH}"
fi

# Determine installation directory
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    if [ "$(id -u)" -eq 0 ]; then
        error "Cannot write to $INSTALL_DIR even as root"
    else
        INSTALL_DIR="${HOME}/.local/bin"
        mkdir -p "$INSTALL_DIR"
        warn "Cannot write to /usr/local/bin. Installing to ${INSTALL_DIR} instead."
        case ":$PATH:" in
            *:"$INSTALL_DIR":*)
                ;;
            *)
                warn "${INSTALL_DIR} is not in your PATH. Please add it to your shell configuration."
                ;;
        esac
    fi
fi

TEMP_BIN="/tmp/easy-ssh-download"
info "Downloading ${BINARY_NAME} (${OS}/${ARCH}) from GitHub..."

# Perform download with curl
if curl -sSfL -H "User-Agent: easy-ssh-installer" -o "$TEMP_BIN" "$DOWNLOAD_URL"; then
    mv "$TEMP_BIN" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    success "Successfully installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}!"
else
    error "Failed to download binary from ${DOWNLOAD_URL}. Please make sure a release exists or compile from source."
fi
