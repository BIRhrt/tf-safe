# Homebrew Tap Setup Guide

This guide explains how to set up a Homebrew tap for tf-safe distribution.

## Step 1: Create Homebrew Tap Repository

1. **Create a new repository** on GitHub named `homebrew-tap`:
   - Go to https://github.com/new
   - Repository name: `homebrew-tap`
   - Description: "Homebrew tap for tf-safe and other tools"
   - Public repository
   - Initialize with README

2. **Clone the repository**:
   ```bash
   git clone https://github.com/BIRhrt/homebrew-tap.git
   cd homebrew-tap
   ```

3. **Create Formula directory**:
   ```bash
   mkdir Formula
   ```

4. **Copy the tf-safe formula**:
   ```bash
   cp ../tf-safe/packaging/homebrew/tf-safe.rb Formula/
   ```

5. **Commit and push**:
   ```bash
   git add Formula/tf-safe.rb
   git commit -m "Add tf-safe formula v1.0.0"
   git push origin main
   ```

## Step 2: Test Installation

1. **Add the tap**:
   ```bash
   brew tap BIRhrt/tap
   ```

2. **Install tf-safe**:
   ```bash
   brew install tf-safe
   ```

3. **Verify installation**:
   ```bash
   tf-safe --version
   ```

## Step 3: Update Formula for Future Releases

When releasing new versions:

1. **Get new checksums**:
   ```bash
   curl -L https://github.com/BIRhrt/tf-safe/releases/download/vX.Y.Z/checksums.txt
   ```

2. **Update formula**:
   - Update version number
   - Update SHA256 checksums
   - Test locally with `brew audit tf-safe`

3. **Commit and push**:
   ```bash
   git add Formula/tf-safe.rb
   git commit -m "Update tf-safe to vX.Y.Z"
   git push origin main
   ```

## Current Formula Status

✅ **Ready for v1.0.0**:
- Formula created with correct SHA256 checksums
- Supports macOS (Intel & Apple Silicon) and Linux (x64 & ARM64)
- Includes shell completion generation
- Basic functionality tests included

## Installation Commands

Once the tap is set up, users can install tf-safe with:

```bash
# Add tap (one-time setup)
brew tap BIRhrt/tap

# Install tf-safe
brew install tf-safe

# Or install directly
brew install BIRhrt/tap/tf-safe
```

## Next Steps

1. Create the `homebrew-tap` repository
2. Set up the formula as described above
3. Test installation on different platforms
4. Consider submitting to homebrew-core for wider distribution