#!/usr/bin/env bash
set -euo pipefail

# ralphkit installer
# Usage: curl -fsSL https://raw.githubusercontent.com/kfroemming/ralphkit/main/install.sh | bash

REPO="kfroemming/ralphkit"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BOLD='\033[1m'
RESET='\033[0m'

info()  { printf "${BOLD}%s${RESET}\n" "$*"; }
ok()    { printf "${GREEN}✓ %s${RESET}\n" "$*"; }
err()   { printf "${RED}✗ %s${RESET}\n" "$*" >&2; }
die()   { err "$*"; exit 1; }

# Detect OS
case "$(uname -s)" in
  Darwin) OS="darwin" ;;
  Linux)  OS="linux" ;;
  *)      die "Unsupported OS: $(uname -s). Only macOS and Linux are supported." ;;
esac

# Detect architecture
case "$(uname -m)" in
  x86_64)       ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)            die "Unsupported architecture: $(uname -m)" ;;
esac

info "Detected ${OS}/${ARCH}"

# Resolve version
if [ -n "${RALPHKIT_VERSION:-}" ]; then
  VERSION="$RALPHKIT_VERSION"
  info "Using specified version: ${VERSION}"
else
  info "Resolving latest version..."
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')
  [ -n "$VERSION" ] || die "Failed to resolve latest version"
  ok "Latest version: ${VERSION}"
fi

# Set up temp dir
TMPDIR_INSTALL=$(mktemp -d)
trap 'rm -rf "$TMPDIR_INSTALL"' EXIT

TARBALL="ralphkit_${VERSION}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/v${VERSION}"

# Download tarball and checksums
info "Downloading ${TARBALL}..."
curl -fsSL -o "${TMPDIR_INSTALL}/${TARBALL}" "${BASE_URL}/${TARBALL}" || die "Download failed"
ok "Downloaded tarball"

info "Downloading checksums..."
curl -fsSL -o "${TMPDIR_INSTALL}/checksums.txt" "${BASE_URL}/checksums.txt" || die "Checksums download failed"
ok "Downloaded checksums"

# Verify checksum
info "Verifying checksum..."
cd "$TMPDIR_INSTALL"
EXPECTED=$(grep "$TARBALL" checksums.txt | awk '{print $1}')
[ -n "$EXPECTED" ] || die "Checksum not found for ${TARBALL}"

if [ "$OS" = "darwin" ]; then
  ACTUAL=$(shasum -a 256 "$TARBALL" | awk '{print $1}')
else
  ACTUAL=$(sha256sum "$TARBALL" | awk '{print $1}')
fi

[ "$EXPECTED" = "$ACTUAL" ] || die "Checksum mismatch!\n  expected: ${EXPECTED}\n  got:      ${ACTUAL}"
ok "Checksum verified"

# Extract
info "Extracting..."
tar xzf "$TARBALL"
ok "Extracted"

# Install
info "Installing to ${INSTALL_DIR}..."
if [ -w "$INSTALL_DIR" ]; then
  mv ralphkit "${INSTALL_DIR}/ralphkit"
else
  info "Requires sudo for ${INSTALL_DIR}"
  sudo mv ralphkit "${INSTALL_DIR}/ralphkit"
fi
chmod +x "${INSTALL_DIR}/ralphkit"
ok "Installed ralphkit to ${INSTALL_DIR}/ralphkit"

# Bootstrap
info "Running ralphkit install to bootstrap dependencies..."
"${INSTALL_DIR}/ralphkit" install || true
ok "Bootstrap complete"

printf "\n${GREEN}${BOLD}ralphkit ${VERSION} installed successfully!${RESET}\n"
printf "Run ${BOLD}ralphkit --help${RESET} to get started.\n"
