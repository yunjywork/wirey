#!/bin/bash

# Wirey macOS Build & Package Script
# Usage: ./scripts/build-macos.sh [version]
# Example: ./scripts/build-macos.sh v1.0.1

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build/bin"
DIST_DIR="$PROJECT_ROOT/dist"

# Version from argument or git describe
VERSION="${1:-$(git describe --tags --always 2>/dev/null || echo "dev")}"

echo -e "${YELLOW}=== Wirey macOS Build Script ===${NC}"
echo -e "Version: ${GREEN}$VERSION${NC}"
echo -e "Project: $PROJECT_ROOT"
echo ""

# Checkout version if tag provided
if [[ "$1" =~ ^v[0-9] ]]; then
    echo -e "${YELLOW}Fetching tags...${NC}"
    git fetch --tags

    echo -e "${YELLOW}Checking out $VERSION...${NC}"
    git checkout "$VERSION"
    echo ""
fi

# Build
echo -e "${YELLOW}Building for darwin/arm64...${NC}"
cd "$PROJECT_ROOT"
wails build -platform darwin/arm64

if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi

echo -e "${GREEN}Build complete!${NC}"
echo ""

# Create dist directory
mkdir -p "$DIST_DIR"

# Create DMG
echo -e "${YELLOW}Creating DMG...${NC}"
cd "$BUILD_DIR"

DMG_NAME="wirey-arm64.dmg"
DMG_PATH="$DIST_DIR/$DMG_NAME"

# Remove old DMG if exists
rm -f "$DMG_PATH"

hdiutil create \
    -volname "Wirey $VERSION" \
    -srcfolder Wirey.app \
    -ov \
    -format UDZO \
    "$DMG_PATH"

if [ $? -ne 0 ]; then
    echo -e "${RED}DMG creation failed!${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}=== Build Complete ===${NC}"
echo -e "Output: ${DMG_PATH}"
ls -lh "$DMG_PATH"
