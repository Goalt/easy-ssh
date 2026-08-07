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
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    i386|i686)
        ARCH="386"
        ;;
    armv7l|armv6l)
        ARCH="arm"
        ;;
    *)
        error "Unsupported architecture: $ARCH"
        ;;
esac

case "$OS" in
    linux)
        OS="linux"
        ;;
    darwin)
        OS="darwin"
        ;;
    *)
        error "Unsupported operating system: $OS"
        ;;
esac

REPO="Goalt/easy-ssh"
BINARY_NAME="easy-ssh"

# Prepare curl auth header if GITHUB_TOKEN or GH_TOKEN is set
CURL_AUTH_HEADER=""
TOKEN="${GITHUB_TOKEN:-${GH_TOKEN}}"
if [ -n "$TOKEN" ]; then
    CURL_AUTH_HEADER="Authorization: Bearer $TOKEN"
fi

curl_fetch() {
    if [ -n "$CURL_AUTH_HEADER" ]; then
        curl -sSf -H "$CURL_AUTH_HEADER" -H "User-Agent: easy-ssh-installer" "$@"
    else
        curl -sSf -H "User-Agent: easy-ssh-installer" "$@"
    fi
}

curl_download() {
    if [ -n "$CURL_AUTH_HEADER" ]; then
        curl -sSfL -H "$CURL_AUTH_HEADER" -H "User-Agent: easy-ssh-installer" "$@"
    else
        curl -sSfL -H "User-Agent: easy-ssh-installer" "$@"
    fi
}

# Fetch latest release tag from GitHub API
info "Fetching latest release version..."
if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
    LATEST_TAG=$(gh release view --repo "$REPO" --json tagName -q .tagName 2>/dev/null || true)
fi

if [ -z "$LATEST_TAG" ]; then
    LATEST_TAG=$(curl_fetch "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | head -n1 | sed -E 's/.*"([^"]+)".*/\1/' | tr -d '[:space:]')
fi

if [ -z "$LATEST_TAG" ] || [ "$LATEST_TAG" = "latest" ]; then
    DOWNLOAD_URL="https://github.com/Goalt/easy-ssh/releases/latest/download/easy-ssh-${OS}-${ARCH}"
else
    DOWNLOAD_URL="https://github.com/Goalt/easy-ssh/releases/download/${LATEST_TAG}/easy-ssh-${OS}-${ARCH}"
fi

# Determine installation directory
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    # If /usr/local/bin is not writeable, check if we can sudo or use user local bin
    if [ "$(id -u)" -eq 0 ]; then
        # Running as root, /usr/local/bin should be writeable, but just in case
        error "Cannot write to $INSTALL_DIR even as root"
    else
        # Not root, try to install to ~/.local/bin
        INSTALL_DIR="${HOME}/.local/bin"
        mkdir -p "$INSTALL_DIR"
        warn "Cannot write to /usr/local/bin. Installing to ${INSTALL_DIR} instead."
        # Check if ~/.local/bin is in PATH
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

# Perform download
DOWNLOAD_SUCCESS=0

if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
    if gh release download "$LATEST_TAG" --repo "$REPO" --pattern "easy-ssh-${OS}-${ARCH}" --output "$TEMP_BIN" --clobber >/dev/null 2>&1; then
        DOWNLOAD_SUCCESS=1
    fi
fi

if [ "$DOWNLOAD_SUCCESS" -eq 0 ]; then
    if curl_download -o "$TEMP_BIN" "$DOWNLOAD_URL"; then
        DOWNLOAD_SUCCESS=1
    fi
fi

if [ "$DOWNLOAD_SUCCESS" -eq 1 ]; then
    mv "$TEMP_BIN" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    success "Successfully installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}!"
else
    error "Failed to download binary. Please make sure a release exists or compile from source."
fi

