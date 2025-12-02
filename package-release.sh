#!/bin/bash

set -e

if [ ! -d "dist" ]; then
    echo "ERROR: dist/ directory not found!"
    echo "Please run build-release.sh first to build the binaries."
    exit 1
fi

echo "Packaging release files..."
echo ""

RELEASE_DIR="release"
rm -rf "$RELEASE_DIR"
mkdir -p "$RELEASE_DIR"

VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
echo "Version: $VERSION"
echo ""

echo "Creating Windows package..."
WIN_DIR="$RELEASE_DIR/dota-report-timestamps-windows"
mkdir -p "$WIN_DIR"
cp dist/windows/*.exe "$WIN_DIR/"
cp launch-windows.bat "$WIN_DIR/"
cp -r cmd/server/static "$WIN_DIR/"
cp -r assets/portraits "$WIN_DIR/assets/"
cp QUICKSTART.txt "$WIN_DIR/README.txt"
cd "$RELEASE_DIR"
zip -r "dota-report-timestamps-windows-$VERSION.zip" "dota-report-timestamps-windows" > /dev/null
cd ..
echo "  Created: release/dota-report-timestamps-windows-$VERSION.zip"

echo "Creating Mac package..."
MAC_DIR="$RELEASE_DIR/dota-report-timestamps-mac"
mkdir -p "$MAC_DIR"
cp dist/mac/bot* "$MAC_DIR/" 2>/dev/null || true
cp dist/mac/server* "$MAC_DIR/" 2>/dev/null || true
cp launch-mac.sh "$MAC_DIR/"
chmod +x "$MAC_DIR/launch-mac.sh"
chmod +x "$MAC_DIR"/bot* 2>/dev/null || true
chmod +x "$MAC_DIR"/server* 2>/dev/null || true
cp -r cmd/server/static "$MAC_DIR/"
cp -r assets/portraits "$MAC_DIR/assets/"
cp QUICKSTART.txt "$MAC_DIR/README.txt"
cd "$RELEASE_DIR"
zip -r "dota-report-timestamps-mac-$VERSION.zip" "dota-report-timestamps-mac" > /dev/null
cd ..
echo "  Created: release/dota-report-timestamps-mac-$VERSION.zip"

echo "Creating Linux package..."
LINUX_DIR="$RELEASE_DIR/dota-report-timestamps-linux"
mkdir -p "$LINUX_DIR"
cp dist/linux/bot "$LINUX_DIR/"
cp dist/linux/server "$LINUX_DIR/"
cp launch-linux.sh "$LINUX_DIR/"
chmod +x "$LINUX_DIR/launch-linux.sh"
chmod +x "$LINUX_DIR/bot"
chmod +x "$LINUX_DIR/server"
cp -r cmd/server/static "$LINUX_DIR/"
cp -r assets/portraits "$LINUX_DIR/assets/"
cp QUICKSTART.txt "$LINUX_DIR/README.txt"
cd "$RELEASE_DIR"
zip -r "dota-report-timestamps-linux-$VERSION.zip" "dota-report-timestamps-linux" > /dev/null
cd ..
echo "  Created: release/dota-report-timestamps-linux-$VERSION.zip"

echo ""
echo "Packaging complete! Release files are in the release/ directory."

