#!/usr/bin/env bash
# Umaru CLI - Linux & macOS Automatic Installer
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Baranigsiz/UmaruCLI/main/install.sh | bash

set -euo pipefail

CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
MAGENTA='\033[0;35m'
GRAY='\033[0;90m'
NC='\033[0m' # No Color

echo ""
echo -e "${MAGENTA}================================================================${NC}"
echo -e "${CYAN}   ⚡ UMARU CLI - Automatic Installer (Linux & macOS)${NC}"
echo -e "${GRAY}   Production Scaffolding in Milliseconds${NC}"
echo -e "${MAGENTA}================================================================${NC}"
echo ""

# 1. Detect Operating System
OS_NAME="$(uname -s)"
case "${OS_NAME}" in
    Linux*)     OS="linux" ;;
    Darwin*)    OS="darwin" ;;
    *)
        echo -e "${RED}[ERROR] Unsupported operating system: ${OS_NAME}${NC}"
        exit 1
        ;;
esac

# 2. Detect Architecture
ARCH_NAME="$(uname -m)"
case "${ARCH_NAME}" in
    x86_64|amd64)   ARCH="amd64" ;;
    arm64|aarch64)  ARCH="arm64" ;;
    *)
        echo -e "${RED}[ERROR] Unsupported system architecture: ${ARCH_NAME}${NC}"
        exit 1
        ;;
esac

echo -e "  ${CYAN}[>] Detected Platform:${NC} ${OS} / ${ARCH}"

# 3. Determine latest version
echo -e "  ${GRAY}[>] Fetching latest release info from GitHub...${NC}"
REPO="Baranigsiz/UmaruCLI"
LATEST_TAG=""

if command -v curl >/dev/null 2>&1; then
    LATEST_TAG=$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || true)
fi

if [ -z "${LATEST_TAG}" ]; then
    LATEST_TAG="v1.9.0"
    echo -e "  ${YELLOW}[!] Could not reach GitHub API, falling back to ${LATEST_TAG}${NC}"
fi

VERSION="${LATEST_TAG#v}"
echo -e "  ${GREEN}[>] Target Version:${NC} ${LATEST_TAG}"

ASSET="umaru_${VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${ASSET}"

# 4. Download and extract
TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TEMP_DIR}"' EXIT

echo -e "  ${CYAN}[>] Downloading ${ASSET}...${NC}"
if command -v curl >/dev/null 2>&1; then
    curl -fsSL "${DOWNLOAD_URL}" -o "${TEMP_DIR}/${ASSET}"
elif command -v wget >/dev/null 2>&1; then
    wget -q "${DOWNLOAD_URL}" -O "${TEMP_DIR}/${ASSET}"
else
    echo -e "${RED}[ERROR] Neither curl nor wget was found on the system.${NC}"
    exit 1
fi

echo -e "  ${GRAY}[>] Extracting archive...${NC}"
tar -xzf "${TEMP_DIR}/${ASSET}" -C "${TEMP_DIR}"

if [ ! -f "${TEMP_DIR}/umaru" ]; then
    echo -e "${RED}[ERROR] Extraction failed: binary 'umaru' not found in archive.${NC}"
    exit 1
fi

chmod +x "${TEMP_DIR}/umaru"

# 5. Determine installation destination
INSTALL_DIR="/usr/local/bin"
USE_SUDO=false

if [ ! -w "${INSTALL_DIR}" ]; then
    if command -v sudo >/dev/null 2>&1 && [ "$EUID" -ne 0 ]; then
        USE_SUDO=true
    else
        INSTALL_DIR="${HOME}/.local/bin"
        mkdir -p "${INSTALL_DIR}"
    fi
fi

TARGET="${INSTALL_DIR}/umaru"
echo -e "  ${GRAY}[>] Installing binary to ${TARGET}...${NC}"

if [ "${USE_SUDO}" = true ]; then
    sudo cp "${TEMP_DIR}/umaru" "${TARGET}"
    sudo chmod +x "${TARGET}"
else
    cp "${TEMP_DIR}/umaru" "${TARGET}"
    chmod +x "${TARGET}"
fi

# 6. Verify and finish
echo ""
echo -e "  ${GREEN}[SUCCESS] Umaru CLI has been installed successfully!${NC}"
echo -e "  ${GRAY}Binary location: ${TARGET}${NC}"
echo ""
echo -e "  ${CYAN}Verification output:${NC}"
"${TARGET}" version || true

# Check if INSTALL_DIR is in PATH
case ":${PATH}:" in
    *:"${INSTALL_DIR}":*) ;;
    *)
        echo ""
        echo -e "  ${YELLOW}[NOTE] Please add ${INSTALL_DIR} to your PATH to run 'umaru' from anywhere:${NC}"
        echo -e "    export PATH=\"${INSTALL_DIR}:\$PATH\""
        ;;
esac

echo ""
echo -e "${MAGENTA}================================================================${NC}"
echo -e "${CYAN}  Get started: Run 'umaru init' to scaffold your first project.${NC}"
echo -e "${YELLOW}  Check system readiness: 'umaru doctor'${NC}"
echo -e "${MAGENTA}================================================================${NC}"
echo ""
