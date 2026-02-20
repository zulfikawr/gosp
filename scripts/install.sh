#!/bin/bash
# GOSP One-Liner Installer 🦞
# This script detects your OS/Arch and downloads the latest GOSP binary.

set -e

# Configuration
REPO="zulfikawr/gosp"
BINARY_NAME="gosp"
INSTALL_DIR="/usr/local/bin"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}🦞 GOSP: Global Open Search Protocol${NC}"
echo -e "Starting installation..."

# 1. Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

echo -e "Detected: ${GREEN}$OS/$ARCH${NC}"

# 2. Get Latest Version Tag from GitHub
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    echo -e "${RED}Error: Could not determine latest version.${NC}"
    exit 1
fi

echo -e "Latest Version: ${GREEN}$LATEST_TAG${NC}"

# 3. Construct Download URL
# Based on gosp/.github/workflows/release.yml naming: gosp-linux-amd64
EXT=""
if [ "$OS" == "windows" ]; then EXT=".exe"; fi
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/${BINARY_NAME}-${OS}-${ARCH}${EXT}"

# 4. Download
TEMP_DIR=$(mktemp -d)
echo -e "Downloading: $DOWNLOAD_URL"

if ! curl -L -o "$TEMP_DIR/$BINARY_NAME" "$DOWNLOAD_URL"; then
    echo -e "${RED}Error: Download failed. Check if a release exists for your OS/Arch.${NC}"
    exit 1
fi

chmod +x "$TEMP_DIR/$BINARY_NAME"

# 5. Move to Install Directory
echo -e "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
else
    sudo mv "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
fi

# 6. Cleanup
rm -rf "$TEMP_DIR"

echo -e "${GREEN}✅ Installation complete!${NC}"
echo -e "Run '${BLUE}gosp --help${NC}' to get started."
