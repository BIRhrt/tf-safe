#!/bin/bash
# Script to build Chocolatey package for tf-safe

set -e

VERSION=${1:-"1.0.0"}
GITHUB_REPO=${2:-"BIRhrt/tf-safe"}
BUILD_DIR="build/chocolatey"
PACKAGE_NAME="tf-safe"

echo "Building Chocolatey package for $PACKAGE_NAME version $VERSION..."

# Check if choco is available (for testing)
if command -v choco >/dev/null 2>&1; then
    CHOCO_AVAILABLE=true
else
    CHOCO_AVAILABLE=false
    echo "Warning: Chocolatey not available for testing"
fi

# Clean and create build directory
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/tools"

# Get checksum for Windows binary
if [ ! -f "dist/tf-safe-windows-amd64.exe" ]; then
    echo "Error: Windows binary not found. Run 'make build-all' first."
    exit 1
fi

# Calculate checksum
CHECKSUM=$(sha256sum "dist/tf-safe-windows-amd64.exe" | cut -d' ' -f1)
echo "Windows binary checksum: $CHECKSUM"

# Copy and update nuspec file
cp packaging/chocolatey/tf-safe.nuspec "$BUILD_DIR/"
# Use a more portable approach for sed
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s/<version>1.0.0<\/version>/<version>$VERSION<\/version>/" "$BUILD_DIR/tf-safe.nuspec"
    sed -i '' "s/your-org\/tf-safe/${GITHUB_REPO//\//\\/}/g" "$BUILD_DIR/tf-safe.nuspec"
else
    sed -i "s/<version>1.0.0<\/version>/<version>$VERSION<\/version>/" "$BUILD_DIR/tf-safe.nuspec"
    sed -i "s/your-org\/tf-safe/${GITHUB_REPO//\//\\/}/g" "$BUILD_DIR/tf-safe.nuspec"
fi

# Copy and update install script
cp packaging/chocolatey/tools/chocolateyinstall.ps1 "$BUILD_DIR/tools/"
cp packaging/chocolatey/tools/chocolateyuninstall.ps1 "$BUILD_DIR/tools/"

# Update install script with correct version and checksum
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s/\\\$version = '1.0.0'/\\\$version = '$VERSION'/" "$BUILD_DIR/tools/chocolateyinstall.ps1"
    sed -i '' "s/your-org\/tf-safe/${GITHUB_REPO//\//\\/}/g" "$BUILD_DIR/tools/chocolateyinstall.ps1"
    sed -i '' "s/REPLACE_WITH_ACTUAL_CHECKSUM/$CHECKSUM/" "$BUILD_DIR/tools/chocolateyinstall.ps1"
else
    sed -i "s/\\\$version = '1.0.0'/\\\$version = '$VERSION'/" "$BUILD_DIR/tools/chocolateyinstall.ps1"
    sed -i "s/your-org\/tf-safe/${GITHUB_REPO//\//\\/}/g" "$BUILD_DIR/tools/chocolateyinstall.ps1"
    sed -i "s/REPLACE_WITH_ACTUAL_CHECKSUM/$CHECKSUM/" "$BUILD_DIR/tools/chocolateyinstall.ps1"
fi

# Create the package
cd "$BUILD_DIR"

if [ "$CHOCO_AVAILABLE" = true ]; then
    echo "Building Chocolatey package..."
    choco pack tf-safe.nuspec
    
    # Move to dist directory
    mkdir -p ../../dist
    mv "${PACKAGE_NAME}.${VERSION}.nupkg" "../../dist/"
    
    echo "Chocolatey package created: dist/${PACKAGE_NAME}.${VERSION}.nupkg"
    echo ""
    echo "To test locally:"
    echo "  choco install tf-safe -s dist/"
    echo ""
    echo "To publish to Chocolatey Community Repository:"
    echo "  choco push dist/${PACKAGE_NAME}.${VERSION}.nupkg -s https://push.chocolatey.org/"
else
    echo "Chocolatey package files prepared in $BUILD_DIR"
    echo "To build the package on Windows:"
    echo "  cd $BUILD_DIR"
    echo "  choco pack tf-safe.nuspec"
    echo ""
    echo "Package files:"
    find . -type f
fi

cd - > /dev/null