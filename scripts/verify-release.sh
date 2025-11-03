#!/bin/bash
# tf-safe Release Verification Script
# Run this after GitHub Actions completes to verify the release

set -e

REPO="BIRhrt/tf-safe"
VERSION="v1.0.0"
RELEASE_URL="https://github.com/${REPO}/releases/tag/${VERSION}"

echo "🔍 Verifying tf-safe ${VERSION} release..."
echo ""

# Check if release exists
echo "1. Checking GitHub release..."
if curl -s "https://api.github.com/repos/${REPO}/releases/tags/${VERSION}" | grep -q "tag_name"; then
    echo "   ✅ Release ${VERSION} found on GitHub"
else
    echo "   ❌ Release ${VERSION} not found"
    exit 1
fi

# Check release assets
echo ""
echo "2. Checking release assets..."
assets=$(curl -s "https://api.github.com/repos/${REPO}/releases/tags/${VERSION}" | grep -o '"name": "[^"]*"' | cut -d'"' -f4)

expected_assets=(
    "tf-safe-linux-amd64.tar.gz"
    "tf-safe-linux-arm64.tar.gz"
    "tf-safe-darwin-amd64.tar.gz"
    "tf-safe-darwin-arm64.tar.gz"
    "tf-safe-windows-amd64.zip"
    "checksums.txt"
)

for asset in "${expected_assets[@]}"; do
    if echo "$assets" | grep -q "$asset"; then
        echo "   ✅ $asset"
    else
        echo "   ❌ $asset (missing)"
    fi
done

# Test installation script
echo ""
echo "3. Testing installation script..."
if curl -fsSL "https://raw.githubusercontent.com/${REPO}/main/scripts/install.sh" | head -1 | grep -q "#!/bin/bash"; then
    echo "   ✅ Installation script accessible"
else
    echo "   ❌ Installation script not accessible"
fi

echo ""
echo "🎉 Release verification complete!"
echo ""
echo "📋 Next steps:"
echo "1. Test installation: curl -fsSL https://raw.githubusercontent.com/${REPO}/main/scripts/install.sh | bash"
echo "2. Create Homebrew tap with checksums from: ${RELEASE_URL}"
echo "3. Announce the release!"
echo ""
echo "🔗 Release URL: ${RELEASE_URL}"