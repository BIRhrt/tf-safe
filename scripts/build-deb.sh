#!/bin/bash
# Script to build DEB packages for tf-safe

set -e

VERSION=${1:-"1.0.0"}
ARCH=${2:-"amd64"}
BUILD_DIR="build/deb"
PACKAGE_NAME="tf-safe"

echo "Building DEB package for $PACKAGE_NAME version $VERSION ($ARCH)..."

# Clean and create build directory
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Create package directory structure
PACKAGE_DIR="$BUILD_DIR/${PACKAGE_NAME}_${VERSION}_${ARCH}"
mkdir -p "$PACKAGE_DIR/DEBIAN"
mkdir -p "$PACKAGE_DIR/usr/local/bin"
mkdir -p "$PACKAGE_DIR/usr/share/doc/$PACKAGE_NAME"

# Copy binary based on architecture
if [ "$ARCH" = "amd64" ]; then
    BINARY_NAME="tf-safe-linux-amd64"
elif [ "$ARCH" = "arm64" ]; then
    BINARY_NAME="tf-safe-linux-arm64"
else
    echo "Error: Unsupported architecture $ARCH"
    exit 1
fi

if [ ! -f "dist/$BINARY_NAME" ]; then
    echo "Error: Binary dist/$BINARY_NAME not found. Run 'make build-all' first."
    exit 1
fi

cp "dist/$BINARY_NAME" "$PACKAGE_DIR/usr/local/bin/tf-safe"
chmod 755 "$PACKAGE_DIR/usr/local/bin/tf-safe"

# Copy control file and update architecture
sed "s/Architecture: amd64/Architecture: $ARCH/" packaging/debian/control > "$PACKAGE_DIR/DEBIAN/control"
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s/Version: 1.0.0/Version: $VERSION/" "$PACKAGE_DIR/DEBIAN/control"
else
    sed -i "s/Version: 1.0.0/Version: $VERSION/" "$PACKAGE_DIR/DEBIAN/control"
fi

# Copy maintainer scripts
cp packaging/debian/postinst "$PACKAGE_DIR/DEBIAN/"
cp packaging/debian/prerm "$PACKAGE_DIR/DEBIAN/"
chmod 755 "$PACKAGE_DIR/DEBIAN/postinst"
chmod 755 "$PACKAGE_DIR/DEBIAN/prerm"

# Create documentation
cat > "$PACKAGE_DIR/usr/share/doc/$PACKAGE_NAME/README.Debian" << EOF
tf-safe for Debian
==================

This package provides the tf-safe CLI tool for Terraform state file protection.

Configuration
-------------

Initialize tf-safe in your Terraform project:

    cd /path/to/terraform/project
    tf-safe init

This will create a .tf-safe.yaml configuration file with default settings.

Usage
-----

Use tf-safe as a wrapper for Terraform commands:

    tf-safe apply
    tf-safe plan
    tf-safe destroy

Or use manual backup/restore commands:

    tf-safe backup
    tf-safe list
    tf-safe restore <backup-id>

For more information, see:
    tf-safe --help
    tf-safe <command> --help

Documentation: https://github.com/your-org/tf-safe
EOF

# Create copyright file
cat > "$PACKAGE_DIR/usr/share/doc/$PACKAGE_NAME/copyright" << EOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: tf-safe
Upstream-Contact: tf-safe Team <maintainers@tf-safe.dev>
Source: https://github.com/your-org/tf-safe

Files: *
Copyright: $(date +%Y) tf-safe Team
License: MIT
 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:
 .
 The above copyright notice and this permission notice shall be included in all
 copies or substantial portions of the Software.
 .
 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 SOFTWARE.
EOF

# Calculate installed size (in KB)
INSTALLED_SIZE=$(du -sk "$PACKAGE_DIR" | cut -f1)
echo "Installed-Size: $INSTALLED_SIZE" >> "$PACKAGE_DIR/DEBIAN/control"

# Build the package
echo "Building DEB package..."
dpkg-deb --build "$PACKAGE_DIR"

# Move to dist directory
mkdir -p dist
mv "$BUILD_DIR/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb" "dist/"

echo "DEB package created: dist/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
echo ""
echo "To install:"
echo "  sudo dpkg -i dist/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
echo ""
echo "To upload to APT repository:"
echo "  1. Sign the package: dpkg-sig --sign builder dist/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
echo "  2. Add to repository with reprepro or similar tool"