#!/bin/sh
set -e

REPO="dsswift/cli-exchange"
BINARY="exchange"

# Detect OS
OS="$(uname -s)"
case "$OS" in
    Linux*)  GOOS="linux" ;;
    Darwin*) GOOS="darwin" ;;
    MINGW*|MSYS*|CYGWIN*) GOOS="windows" ;;
    *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)  GOARCH="amd64" ;;
    arm64|aarch64) GOARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Check for supported platform combinations
case "${GOOS}-${GOARCH}" in
    linux-amd64|darwin-amd64|darwin-arm64|windows-amd64) ;;
    *) echo "No release binary for ${GOOS}/${GOARCH}" >&2; exit 1 ;;
esac

# Resolve latest release tag
if command -v curl >/dev/null 2>&1; then
    TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | cut -d'"' -f4)
elif command -v wget >/dev/null 2>&1; then
    TAG=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | cut -d'"' -f4)
else
    echo "curl or wget is required" >&2
    exit 1
fi

if [ -z "$TAG" ]; then
    echo "Failed to resolve latest release" >&2
    exit 1
fi

# Build download URL
EXT=""
if [ "$GOOS" = "windows" ]; then EXT=".exe"; fi
ASSET="${BINARY}-${GOOS}-${GOARCH}${EXT}"
URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

# Set install directory
if [ "$GOOS" = "windows" ]; then
    INSTALL_DIR="${LOCALAPPDATA:-$HOME/AppData/Local}/Programs/exchange"
else
    INSTALL_DIR="${HOME}/.local/bin"
fi
mkdir -p "$INSTALL_DIR"

DEST="${INSTALL_DIR}/${BINARY}${EXT}"

echo "Installing ${BINARY} ${TAG} (${GOOS}/${GOARCH})..."

# Download
TMPFILE=$(mktemp)
trap 'rm -f "$TMPFILE"' EXIT
if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$TMPFILE" "$URL"
else
    wget -qO "$TMPFILE" "$URL"
fi

# Verify checksum
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${TAG}/checksums.txt"
CHECKSUMS=$(mktemp)
trap 'rm -f "$TMPFILE" "$CHECKSUMS"' EXIT
if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$CHECKSUMS" "$CHECKSUMS_URL"
else
    wget -qO "$CHECKSUMS" "$CHECKSUMS_URL"
fi

EXPECTED=$(grep "$ASSET" "$CHECKSUMS" | awk '{print $1}')
if [ -n "$EXPECTED" ]; then
    if command -v sha256sum >/dev/null 2>&1; then
        ACTUAL=$(sha256sum "$TMPFILE" | awk '{print $1}')
    elif command -v shasum >/dev/null 2>&1; then
        ACTUAL=$(shasum -a 256 "$TMPFILE" | awk '{print $1}')
    else
        ACTUAL=""
    fi

    if [ -n "$ACTUAL" ] && [ "$ACTUAL" != "$EXPECTED" ]; then
        echo "Checksum mismatch" >&2
        echo "  expected: $EXPECTED" >&2
        echo "  got:      $ACTUAL" >&2
        exit 1
    fi
fi

# Install
mv "$TMPFILE" "$DEST"
chmod +x "$DEST"

echo "Installed to $DEST"

# Check PATH
case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
        echo ""
        echo "Add ${INSTALL_DIR} to your PATH:"
        echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        ;;
esac
