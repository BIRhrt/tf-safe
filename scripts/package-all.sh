#!/bin/bash
# Script to build all installation packages for tf-safe

set -e

VERSION=${1:-"1.0.0"}
GITHUB_REPO=${2:-"your-org/tf-safe"}

echo "Building all packages for tf-safe version $VERSION"
echo "Repository: $GITHUB_REPO"
echo

# Ensure we have built binaries
if [ ! -d "dist" ] || [ -z "$(ls -A dist/tf-safe-* 2>/dev/null)" ]; then
    echo "Building binaries first..."
    make build-all
fi

echo "Building packages..."

# Build DEB packages
echo
echo "=== Building DEB packages ==="
if command -v dpkg-deb >/dev/null 2>&1; then
    ./scripts/build-deb.sh "$VERSION" "amd64"
    ./scripts/build-deb.sh "$VERSION" "arm64"
    echo "DEB packages built successfully"
else
    echo "Warning: dpkg-deb not available, skipping DEB package creation"
    echo "DEB package files prepared in packaging/debian/"
fi

# Build Chocolatey package
echo
echo "=== Building Chocolatey package ==="
./scripts/build-chocolatey.sh "$VERSION" "$GITHUB_REPO"

# Update Homebrew formula (requires release to be published)
echo
echo "=== Preparing Homebrew formula ==="
if curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/tags/v$VERSION" | grep -q "tag_name"; then
    echo "Release v$VERSION found, updating Homebrew formula..."
    ./scripts/update-homebrew-formula.sh "$VERSION" "$GITHUB_REPO"
else
    echo "Release v$VERSION not found on GitHub"
    echo "Homebrew formula template available in packaging/homebrew/"
    echo "Run './scripts/update-homebrew-formula.sh $VERSION $GITHUB_REPO' after publishing the release"
fi

echo
echo "=== Package Summary ==="
echo "Built packages:"

# List created packages
if [ -d "dist" ]; then
    echo
    echo "Binaries:"
    ls -la dist/tf-safe-* 2>/dev/null || echo "  No binaries found"
    
    echo
    echo "DEB packages:"
    ls -la dist/*.deb 2>/dev/null || echo "  No DEB packages found"
    
    echo
    echo "Chocolatey packages:"
    ls -la dist/*.nupkg 2>/dev/null || echo "  No Chocolatey packages found"
fi

echo
echo "Package files prepared:"
echo "  Homebrew formula: packaging/homebrew/tf-safe.rb"
echo "  DEB control files: packaging/debian/"
echo "  Chocolatey files: packaging/chocolatey/"

echo
echo "=== Distribution Instructions ==="
echo
echo "1. Homebrew:"
echo "   - Submit packaging/homebrew/tf-safe.rb to homebrew-core"
echo "   - Or create a custom tap with the formula"
echo
echo "2. APT Repository:"
echo "   - Sign DEB packages: dpkg-sig --sign builder dist/*.deb"
echo "   - Add to repository with reprepro or similar"
echo
echo "3. Chocolatey:"
echo "   - Test locally: choco install tf-safe -s dist/"
echo "   - Publish: choco push dist/tf-safe.$VERSION.nupkg"
echo
echo "4. Manual Installation:"
echo "   - Upload binaries to GitHub releases"
echo "   - Update INSTALL.md with download links"

echo
echo "Package creation completed!"