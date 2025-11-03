# 🎉 tf-safe v1.0.0 Released!

We're excited to announce the first stable release of **tf-safe** - a lightweight CLI tool for comprehensive Terraform state file protection!

## 🚀 What is tf-safe?

tf-safe provides automated backup, encryption, versioning, and restoration capabilities for your Terraform state files. Never lose your Terraform state again!

### ✨ Key Features

- **🔄 Automated Backups**: Automatic state backups before and after Terraform operations
- **☁️ Multiple Storage Backends**: Local filesystem and AWS S3 support
- **🔐 Encryption**: AES-256-GCM and AWS KMS encryption options
- **🔧 Terraform Integration**: Drop-in replacement for terraform commands
- **⚙️ Flexible Configuration**: Project-level and global configuration support
- **🌍 Cross-Platform**: Single binary for Linux, macOS, and Windows
- **📦 Retention Policies**: Configurable backup retention with automatic cleanup
- **🔄 State Restoration**: Easy restoration of any previous state version

## 📦 Installation

### Quick Install Script
```bash
curl -fsSL https://raw.githubusercontent.com/BIRhrt/tf-safe/main/scripts/install.sh | bash
```

### Homebrew (macOS/Linux)
```bash
brew tap BIRhrt/tap
brew install tf-safe
```

### Manual Installation
Download binaries from [GitHub Releases](https://github.com/BIRhrt/tf-safe/releases/tag/v1.0.0)

## 🏃 Quick Start

1. **Initialize tf-safe in your Terraform project**:
   ```bash
   tf-safe init
   ```

2. **Configure your preferred storage and encryption**:
   ```bash
   # Edit .tf-safe.yaml with your preferences
   ```

3. **Use tf-safe instead of terraform**:
   ```bash
   tf-safe apply    # Automatically creates backups
   tf-safe plan     # Safe planning with backup hooks
   tf-safe destroy  # Backup before destruction
   ```

4. **Manage your backups**:
   ```bash
   tf-safe list                    # List all backups
   tf-safe backup                  # Manual backup
   tf-safe restore <backup-id>     # Restore from backup
   ```

## 🔧 Configuration Example

```yaml
# .tf-safe.yaml
local:
  enabled: true
  path: ".tfstate_snapshots"
  retention_count: 10

remote:
  enabled: true
  provider: "s3"
  bucket: "my-terraform-backups"
  region: "us-west-2"
  retention_count: 50

encryption:
  provider: "aes"
  passphrase: "your-secure-passphrase"
```

## 🌟 Why tf-safe?

- **Peace of Mind**: Never worry about losing Terraform state again
- **Zero Configuration**: Works out of the box with sensible defaults
- **Production Ready**: Comprehensive testing and error handling
- **Developer Friendly**: Seamless integration with existing workflows
- **Secure**: Multiple encryption options for sensitive state data

## 📊 Platform Support

| Platform | Architecture | Status |
|----------|--------------|--------|
| Linux | x64 | ✅ |
| Linux | ARM64 | ✅ |
| macOS | Intel | ✅ |
| macOS | Apple Silicon | ✅ |
| Windows | x64 | ✅ |

## 🔗 Links

- **GitHub Repository**: https://github.com/BIRhrt/tf-safe
- **Documentation**: https://github.com/BIRhrt/tf-safe/blob/main/README.md
- **Issues & Support**: https://github.com/BIRhrt/tf-safe/issues
- **Releases**: https://github.com/BIRhrt/tf-safe/releases

## 🤝 Contributing

We welcome contributions! Check out our [Contributing Guide](https://github.com/BIRhrt/tf-safe/blob/main/CONTRIBUTING.md) to get started.

## 📄 License

tf-safe is released under the [MIT License](https://github.com/BIRhrt/tf-safe/blob/main/LICENSE).

---

**Try tf-safe today and protect your Terraform infrastructure!** 🛡️

*Have questions or feedback? Open an issue or start a discussion on GitHub!*