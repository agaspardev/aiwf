#!/bin/sh
# aiwf — install script (Linux/macOS)
# Usage: curl -sfL https://github.com/agaspardev/aiwf/releases/latest/download/install.sh | sh

set -eu

REPO="agaspardev/aiwf"
BINARY="aiwf"

# Detect OS and architecture
detect_os() {
  case "$(uname -s)" in
    Linux)  echo "linux" ;;
    Darwin) echo "darwin" ;;
    *)      echo "unsupported: $(uname -s)"; exit 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *)             echo "unsupported: $(uname -m)"; exit 1 ;;
  esac
}

OS=$(detect_os)
ARCH=$(detect_arch)

# Detect latest version from GitHub Releases
echo "🔍 Detectando última versión..."
VERSION=$(curl -sSfL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
if [ -z "$VERSION" ]; then
  echo "❌ No se pudo detectar la última versión"
  exit 1
fi
echo "📦 Versión: $VERSION"

# Download archive
ARCHIVE="aiwf_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE"
echo "⬇️ Descargando $URL..."
curl -sSfL -o "/tmp/$ARCHIVE" "$URL"

# Extract
echo "📂 Extrayendo..."
tar -xzf "/tmp/$ARCHIVE" -C /tmp "$BINARY"

# Install
INSTALL_DIR="${AIWF_INSTALL_DIR:-/usr/local/bin}"
echo "📁 Instalando en $INSTALL_DIR..."
sudo mkdir -p "$INSTALL_DIR"
sudo mv "/tmp/$BINARY" "$INSTALL_DIR/$BINARY"
sudo chmod +x "$INSTALL_DIR/$BINARY"

# Cleanup
rm -f "/tmp/$ARCHIVE"

echo "✅ aiwf instalado en $INSTALL_DIR/$BINARY"
echo "   Ejecuta 'aiwf doctor' para verificar"
