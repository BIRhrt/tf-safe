# tf-safe Publication Quick Checklist

## 🚀 Ready to Publish? Follow These Steps

### ✅ Pre-Flight Check (5 minutes)

```bash
# 1. Clean and test everything
make clean
make test
make lint
make build-all

# 2. Verify all files are present
ls -la LICENSE CONTRIBUTING.md CHANGELOG.md README.md
ls -la docs/ examples/

# 3. Check version consistency
grep -r "version" main.go README.md
```

### 🏗️ Repository Setup (10 minutes)

1. **Create GitHub repository**:
   - Go to https://github.com/new
   - Name: `tf-safe`
   - Description: "Lightweight CLI tool for Terraform state file protection"
   - Public repository
   - Don't initialize with README (we have our own)

2. **Push code**:
   ```bash
   git remote add origin https://github.com/YOUR_USERNAME/tf-safe.git
   git branch -M main
   git push -u origin main
   ```

3. **Configure repository**:
   - Enable Issues and Discussions
   - Add topics: `terraform`, `backup`, `cli`, `golang`, `devops`

### 🔧 CI/CD Setup (5 minutes)

1. **Add repository secrets** (Settings → Secrets and variables → Actions):
   ```
   AWS_ACCESS_KEY_ID (for testing S3 functionality)
   AWS_SECRET_ACCESS_KEY (for testing S3 functionality)
   ```

2. **Test CI pipeline**:
   ```bash
   # Create test PR to verify CI works
   git checkout -b test/ci
   echo "# CI Test" >> CI_TEST.md
   git add CI_TEST.md
   git commit -m "test: verify CI pipeline"
   git push origin test/ci
   # Create PR on GitHub and verify builds pass
   ```

### 🏷️ First Release (5 minutes)

1. **Update version and changelog**:
   ```bash
   # Edit CHANGELOG.md - add v1.0.0 section
   # Commit changes
   git add CHANGELOG.md
   git commit -m "chore: prepare v1.0.0 release"
   git push origin main
   ```

2. **Create release tag**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0

   🎉 Initial stable release of tf-safe!

   Features:
   - Automated Terraform state backups
   - Local and S3 storage backends  
   - AES and KMS encryption
   - Cross-platform support (Linux, macOS, Windows)
   - Comprehensive documentation and examples

   Installation:
   - Download binaries from GitHub releases
   - Use installation script: curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/tf-safe/main/scripts/install.sh | bash"

   git push origin v1.0.0
   ```

3. **Monitor release**:
   - Go to GitHub Actions tab
   - Watch release workflow complete
   - Verify binaries are attached to release

### 📦 Package Distribution (15 minutes)

#### Homebrew (Recommended first)

1. **Create Homebrew tap**:
   ```bash
   # Create new repository: homebrew-tap
   git clone https://github.com/YOUR_USERNAME/homebrew-tap.git
   cd homebrew-tap
   mkdir Formula
   ```

2. **Update formula with your details**:
   ```bash
   # Copy and edit the formula
   cp ../tf-safe/packaging/homebrew/tf-safe.rb Formula/
   
   # Edit Formula/tf-safe.rb:
   # - Update homepage URL
   # - Update download URLs with your GitHub username
   # - Update SHA256 checksums from GitHub release
   
   git add Formula/tf-safe.rb
   git commit -m "Add tf-safe formula"
   git push origin main
   ```

3. **Test installation**:
   ```bash
   brew tap YOUR_USERNAME/tap
   brew install tf-safe
   tf-safe --version
   ```

#### Manual Installation Script

1. **Update install script**:
   ```bash
   # Edit scripts/install.sh
   # Replace YOUR_USERNAME with your GitHub username
   sed -i 's/your-org/YOUR_USERNAME/g' scripts/install.sh
   
   git add scripts/install.sh
   git commit -m "fix: update install script with correct repository"
   git push origin main
   ```

2. **Test installation**:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/tf-safe/main/scripts/install.sh | bash
   ```

### 📢 Announce Release (10 minutes)

1. **Update README badges**:
   ```bash
   # Edit README.md to update badge URLs with your username
   # Update build status, license, and other badges
   ```

2. **Create announcement**:
   - GitHub Discussions → Announcements
   - Title: "🎉 tf-safe v1.0.0 Released!"
   - Share key features and installation instructions

3. **Social media** (optional):
   - Twitter/X post about the release
   - LinkedIn post in relevant groups
   - Reddit posts in r/Terraform, r/devops

### ✅ Post-Publication (5 minutes)

1. **Test everything works**:
   ```bash
   # Test Homebrew installation
   brew install YOUR_USERNAME/tap/tf-safe
   
   # Test manual installation
   curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/tf-safe/main/scripts/install.sh | bash
   
   # Test basic functionality
   tf-safe --version
   tf-safe --help
   ```

2. **Monitor and respond**:
   - Watch GitHub Issues for bug reports
   - Respond to questions in Discussions
   - Monitor download statistics

## 🎯 Quick Commands Reference

```bash
# Build and test
make clean && make test && make build-all

# Create release
git tag -a v1.0.0 -m "Release v1.0.0" && git push origin v1.0.0

# Test installation
curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/tf-safe/main/scripts/install.sh | bash

# Update Homebrew formula
# 1. Get SHA256 from GitHub release
# 2. Update Formula/tf-safe.rb with new URLs and checksums
# 3. git add . && git commit -m "Update tf-safe to v1.0.0" && git push
```

## 🚨 Troubleshooting

### Release Build Fails
- Check GitHub Actions logs
- Verify Go version compatibility
- Ensure all tests pass locally

### Installation Script Fails
- Check file permissions
- Verify download URLs are correct
- Test with different operating systems

### Homebrew Formula Issues
- Verify SHA256 checksums match
- Check formula syntax with `brew audit`
- Test installation in clean environment

## 📈 Success Indicators

After 24 hours, you should see:
- ✅ GitHub release with download counts > 0
- ✅ Successful installations via script and Homebrew
- ✅ No critical issues reported
- ✅ Basic community engagement (stars, discussions)

## 🎉 You're Published!

Once you complete these steps, tf-safe will be:
- ✅ Available on GitHub with automated releases
- ✅ Installable via Homebrew and install script
- ✅ Documented with examples and guides
- ✅ Ready for community contributions

**Next steps**: Monitor usage, respond to feedback, and plan future features based on user needs.

---

**Estimated total time**: 45-60 minutes for complete publication

**Need help?** Check the detailed [PUBLISHING.md](PUBLISHING.md) guide or create an issue for assistance.