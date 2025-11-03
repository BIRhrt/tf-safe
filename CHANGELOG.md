# Changelog

All notable changes to tf-safe will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Future enhancements will be listed here

## [1.0.0] - 2025-11-04

### Added
- Initial stable release of tf-safe
- Automated Terraform state backup functionality
- Local filesystem storage backend
- AWS S3 remote storage backend
- AES-256-GCM encryption support
- AWS KMS encryption support
- Terraform command wrappers (apply, plan, destroy)
- Manual backup and restore commands
- Configurable retention policies
- Cross-platform support (Linux, macOS, Windows)
- Comprehensive configuration system
- CLI with Cobra framework
- Integration tests
- Comprehensive documentation
- Example configurations and workflows
- CI/CD pipeline examples (GitHub Actions, GitLab CI, Azure DevOps)
- Package distribution (Homebrew, Chocolatey, DEB)

### Security
- Encryption at rest for all backups
- Secure key management
- Checksum verification for backup integrity

---

## Release Notes Template

When creating a new release, use this template:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features

### Changed
- Changes in existing functionality

### Deprecated
- Soon-to-be removed features

### Removed
- Now removed features

### Fixed
- Bug fixes

### Security
- Security improvements
```

### Categories

- **Added** for new features
- **Changed** for changes in existing functionality
- **Deprecated** for soon-to-be removed features
- **Removed** for now removed features
- **Fixed** for any bug fixes
- **Security** for security improvements