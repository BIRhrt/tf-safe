#!/bin/bash
# Script to update Homebrew formula with correct version and checksums

set -e

VERSION=${1:-"1.0.0"}
GITHUB_REPO=${2:-"your-org/tf-safe"}
FORMULA_FILE="packaging/homebrew/tf-safe.rb"

if [ ! -f "$FORMULA_FILE" ]; then
    echo "Error: Formula file not found at $FORMULA_FILE"
    exit 1
fi

echo "Updating Homebrew formula for version $VERSION..."

# Download checksums file from GitHub release
CHECKSUMS_URL="https://github.com/$GITHUB_REPO/releases/download/v$VERSION/checksums.txt"
echo "Downloading checksums from $CHECKSUMS_URL..."

if ! curl -sL "$CHECKSUMS_URL" -o /tmp/checksums.txt; then
    echo "Error: Could not download checksums file. Make sure the release exists."
    exit 1
fi

# Extract checksums for each platform
DARWIN_AMD64_SHA=$(grep "tf-safe-darwin-amd64.tar.gz" /tmp/checksums.txt | cut -d' ' -f1)
DARWIN_ARM64_SHA=$(grep "tf-safe-darwin-arm64.tar.gz" /tmp/checksums.txt | cut -d' ' -f1)
LINUX_AMD64_SHA=$(grep "tf-safe-linux-amd64.tar.gz" /tmp/checksums.txt | cut -d' ' -f1)
LINUX_ARM64_SHA=$(grep "tf-safe-linux-arm64.tar.gz" /tmp/checksums.txt | cut -d' ' -f1)

if [ -z "$DARWIN_AMD64_SHA" ] || [ -z "$DARWIN_ARM64_SHA" ] || [ -z "$LINUX_AMD64_SHA" ] || [ -z "$LINUX_ARM64_SHA" ]; then
    echo "Error: Could not extract all required checksums"
    cat /tmp/checksums.txt
    exit 1
fi

echo "Extracted checksums:"
echo "  Darwin AMD64: $DARWIN_AMD64_SHA"
echo "  Darwin ARM64: $DARWIN_ARM64_SHA"
echo "  Linux AMD64:  $LINUX_AMD64_SHA"
echo "  Linux ARM64:  $LINUX_ARM64_SHA"

# Create updated formula
cat > "$FORMULA_FILE" << EOF
# Homebrew Formula for tf-safe
# This file should be submitted to homebrew-core or maintained in a custom tap

class TfSafe < Formula
  desc "Terraform state file protection and backup tool"
  homepage "https://github.com/$GITHUB_REPO"
  version "$VERSION"
  
  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/$GITHUB_REPO/releases/download/v#{version}/tf-safe-darwin-arm64.tar.gz"
      sha256 "$DARWIN_ARM64_SHA"
    else
      url "https://github.com/$GITHUB_REPO/releases/download/v#{version}/tf-safe-darwin-amd64.tar.gz"
      sha256 "$DARWIN_AMD64_SHA"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/$GITHUB_REPO/releases/download/v#{version}/tf-safe-linux-arm64.tar.gz"
      sha256 "$LINUX_ARM64_SHA"
    else
      url "https://github.com/$GITHUB_REPO/releases/download/v#{version}/tf-safe-linux-amd64.tar.gz"
      sha256 "$LINUX_AMD64_SHA"
    end
  end

  depends_on "terraform" => :recommended

  def install
    bin.install "tf-safe-#{OS.kernel_name.downcase}-#{Hardware::CPU.arch}" => "tf-safe"
    
    # Generate shell completions
    generate_completions_from_executable(bin/"tf-safe", "completion")
  end

  test do
    system "#{bin}/tf-safe", "--version"
    
    # Test basic functionality
    system "#{bin}/tf-safe", "--help"
    
    # Test init command (should create config file)
    system "#{bin}/tf-safe", "init", "--dry-run"
  end
end
EOF

echo "Updated $FORMULA_FILE with version $VERSION and checksums"
echo ""
echo "To submit to Homebrew:"
echo "1. Fork https://github.com/Homebrew/homebrew-core"
echo "2. Copy $FORMULA_FILE to Formula/tf-safe.rb in your fork"
echo "3. Create a pull request"
echo ""
echo "Or create a custom tap:"
echo "1. Create a repository named homebrew-tap"
echo "2. Add $FORMULA_FILE as Formula/tf-safe.rb"
echo "3. Users can install with: brew install your-org/tap/tf-safe"

# Cleanup
rm -f /tmp/checksums.txt