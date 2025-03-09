#!/bin/bash
set -e

# Custom binary name
BINARY_NAME="httpprobe"
# Set default installation directory
INSTALL_DIR="/usr/local/bin"
# GitHub repository details
REPO_OWNER="mrfoh"
REPO_NAME="httpprobe"

# Print header
echo "=== HttpProbe Installer ==="
echo "This script will download and install the latest version of HttpProbe."

# Determine platform and architecture
detect_platform() {
  PLATFORM="$(uname -s | tr '[:upper:]' '[:lower:]')"
  
  case "$PLATFORM" in
    linux) PLATFORM="linux" ;;
    darwin) PLATFORM="darwin" ;;
    mingw*|msys*|cygwin*) PLATFORM="windows" ;;
    *)
      echo "Unsupported platform: $PLATFORM"
      exit 1
      ;;
  esac
}

detect_arch() {
  ARCH="$(uname -m)"
  
  case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    i386|i686) ARCH="386" ;;
    *)
      echo "Unsupported architecture: $ARCH"
      exit 1
      ;;
  esac
}

# Get the latest release version from GitHub
get_latest_version() {
  if command -v curl &> /dev/null; then
    LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  elif command -v wget &> /dev/null; then
    LATEST_RELEASE=$(wget -q -O - "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  else
    echo "Error: curl or wget is required to download the binary."
    exit 1
  fi
  
  if [ -z "$LATEST_RELEASE" ]; then
    echo "Error: Could not determine the latest release version."
    exit 1
  fi
  
  # Remove leading 'v' if present
  VERSION=${LATEST_RELEASE#v}
}

# Download and install the binary
download_and_install() {
  BINARY_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_RELEASE/${BINARY_NAME}-${VERSION}_${PLATFORM}_${ARCH}.tar.gz"
  TMP_DIR=$(mktemp -d)
  TAR_FILE="$TMP_DIR/${BINARY_NAME}.tar.gz"
  
  echo "Downloading HttpProbe $VERSION for $PLATFORM/$ARCH..."
  echo "URL: $BINARY_URL"
  
  # Download the binary
  if command -v curl &> /dev/null; then
    curl -SL "$BINARY_URL" -o "$TAR_FILE"
  elif command -v wget &> /dev/null; then
    wget -q "$BINARY_URL" -O "$TAR_FILE"
  fi
  
  if [ ! -f "$TAR_FILE" ]; then
    echo "Error: Failed to download the binary."
    exit 1
  fi
  
  # Extract the binary
  echo "Extracting..."
  tar -xzf "$TAR_FILE" -C "$TMP_DIR"
  
  # Find the binary in the extracted files
  BINARY_PATH=$(find "$TMP_DIR" -name "$BINARY_NAME*" -type f | head -n 1)
  
  if [ -z "$BINARY_PATH" ]; then
    echo "Error: Could not find the binary in the downloaded archive."
    exit 1
  fi
  
  # Make the binary executable
  chmod +x "$BINARY_PATH"
  
  # Check if INSTALL_DIR exists and is in PATH
  if [ ! -d "$INSTALL_DIR" ]; then
    echo "Installation directory $INSTALL_DIR does not exist. Creating it now..."
    mkdir -p "$INSTALL_DIR" || sudo mkdir -p "$INSTALL_DIR"
  fi
  
  # Install the binary
  echo "Installing to $INSTALL_DIR/$BINARY_NAME..."
  if [ -w "$INSTALL_DIR" ]; then
    cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
  else
    sudo cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
  fi
  
  # Clean up
  rm -rf "$TMP_DIR"
  
  # Verify installation
  if command -v "$BINARY_NAME" &> /dev/null; then
    echo "Installation successful! HttpProbe $VERSION has been installed to $INSTALL_DIR/$BINARY_NAME"
    echo "Run '$BINARY_NAME --help' to get started."
  else
    echo "Installation may have succeeded, but $BINARY_NAME is not in your PATH."
    echo "The binary was installed to: $INSTALL_DIR/$BINARY_NAME"
    echo "Make sure $INSTALL_DIR is in your PATH, or move the binary to a directory in your PATH."
  fi
}

# Main installation flow
detect_platform
detect_arch
get_latest_version
download_and_install

exit 0