# tf-safe

[![Build Status](https://github.com/your-org/tf-safe/workflows/build/badge.svg)](https://github.com/your-org/tf-safe/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/your-org/tf-safe)](https://goreportcard.com/report/github.com/your-org/tf-safe)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight CLI tool that provides comprehensive Terraform state file protection through automated backup, encryption, versioning, and restoration capabilities. Never lose your Terraform state again!

## 🚀 Features

- **Automated Backups**: Automatic state backups before and after Terraform operations
- **Multiple Storage Backends**: Local filesystem and AWS S3 support
- **Encryption**: AES-256-GCM and AWS KMS encryption options
- **Terraform Integration**: Drop-in replacement for terraform commands
- **Flexible Configuration**: Project-level and global configuration support
- **Cross-Platform**: Single binary for Linux, macOS, and Windows
- **Retention Policies**: Configurable backup retention with automatic cleanup
- **State Restoration**: Easy restoration of any previous state version

## 📦 Installation

### Quick Install

#### macOS (Homebrew)
```bash
brew install tf-safe
```

#### Linux (APT - Debian/Ubuntu)
```bash
curl -fsSL https://packages.tf-safe.dev/gpg | sudo apt-key add -
echo "deb https://packages.tf-safe.dev/apt stable main" | sudo tee /etc/apt/sources.list.d/tf-safe.list
sudo apt update && sudo apt install tf-safe
```

#### Windows (Chocolatey)
```powershell
choco install tf-safe
```

#### Universal Install Script
```bash
curl -fsSL https://raw.githubusercontent.com/your-org/tf-safe/main/scripts/install.sh | bash
```

For detailed installation instructions, see [INSTALL.md](INSTALL.md).

## 🏃 Quick Start

1. **Initialize tf-safe in your Terraform project**:
   ```bash
   cd /path/to/terraform/project
   tf-safe init
   ```

2. **Configure your storage backend** (edit `.tf-safe.yaml`):
   ```yaml
   local:
     enabled: true
     retention_count: 10
   
   remote:
     provider: s3
     bucket: my-terraform-backups
     region: us-west-2
     enabled: true
   
   encryption:
     provider: aes
   ```

3. **Replace terraform commands with tf-safe**:
   ```bash
   # Instead of: terraform apply
   tf-safe apply
   
   # Instead of: terraform plan
   tf-safe plan
   
   # Instead of: terraform destroy
   tf-safe destroy
   ```

4. **View and restore backups**:
   ```bash
   # List all backups
   tf-safe list
   
   # Restore a specific backup
   tf-safe restore 2024-01-15T10:30:00Z
   ```

## 📖 Usage

### Commands

#### `tf-safe init`
Initialize tf-safe configuration in the current directory.

```bash
tf-safe init [flags]

Flags:
  --global          Create global configuration instead of project-level
  --storage string  Storage backend (local, s3) (default "local")
  --encrypt         Enable encryption (default true)
```

#### `tf-safe backup`
Create a manual backup of the current Terraform state.

```bash
tf-safe backup [flags]

Flags:
  --message string  Optional backup message/description
  --force          Force backup even if no changes detected
```

#### `tf-safe list`
List all available backups.

```bash
tf-safe list [flags]

Flags:
  --storage string  Filter by storage backend (local, s3, all) (default "all")
  --limit int      Limit number of results (default 20)
  --format string  Output format (table, json, yaml) (default "table")
```

#### `tf-safe restore`
Restore a previous state backup.

```bash
tf-safe restore <backup-id> [flags]

Flags:
  --dry-run        Show what would be restored without making changes
  --force          Skip confirmation prompt
  --backup-current Create backup of current state before restore (default true)
```

#### Terraform Wrapper Commands
tf-safe provides drop-in replacements for common Terraform commands:

```bash
# Terraform apply with automatic backups
tf-safe apply [terraform-flags]

# Terraform plan with pre-operation backup
tf-safe plan [terraform-flags]

# Terraform destroy with pre-operation backup
tf-safe destroy [terraform-flags]
```

All Terraform flags and arguments are passed through unchanged.

### Configuration

tf-safe uses a hierarchical configuration system:

1. **Command-line flags** (highest priority)
2. **Project-level** `.tf-safe.yaml`
3. **Global** `~/.tf-safe/config.yaml`
4. **Built-in defaults** (lowest priority)

#### Configuration File Structure

```yaml
# Local storage configuration
local:
  enabled: true                    # Enable local backups
  path: ".tfstate_snapshots"      # Local backup directory
  retention_count: 10             # Number of local backups to keep

# Remote storage configuration
remote:
  provider: "s3"                  # Storage provider (s3)
  bucket: "my-terraform-backups"  # S3 bucket name
  region: "us-west-2"            # AWS region
  prefix: "project-name/"        # S3 key prefix
  enabled: true                  # Enable remote backups

# Encryption configuration
encryption:
  provider: "aes"                # Encryption provider (aes, kms, none)
  kms_key_id: ""                # AWS KMS key ID (for KMS provider)

# Retention policies
retention:
  local_count: 10               # Local backup retention count
  remote_count: 50              # Remote backup retention count
  max_age_days: 90             # Maximum backup age in days

# Logging configuration
logging:
  level: "info"                 # Log level (debug, info, warn, error)
  format: "text"               # Log format (text, json)
```

For complete configuration reference, see [Configuration Reference](#configuration-reference).

## 🔧 Advanced Usage

### Multiple Storage Backends

tf-safe can store backups to multiple backends simultaneously:

```yaml
local:
  enabled: true
  retention_count: 5

remote:
  provider: s3
  bucket: team-terraform-backups
  enabled: true
```

### Encryption Options

#### AES-256-GCM (Local Key)
```yaml
encryption:
  provider: aes
  # Key will be generated and stored locally
```

#### AWS KMS
```yaml
encryption:
  provider: kms
  kms_key_id: "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012"
```

#### No Encryption
```yaml
encryption:
  provider: none
```

### CI/CD Integration

#### GitHub Actions
```yaml
name: Terraform with tf-safe
on: [push, pull_request]

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install tf-safe
        run: |
          curl -fsSL https://raw.githubusercontent.com/your-org/tf-safe/main/scripts/install.sh | bash
      
      - name: Configure tf-safe
        run: |
          tf-safe init --storage s3
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
      
      - name: Terraform Plan
        run: tf-safe plan
      
      - name: Terraform Apply
        if: github.ref == 'refs/heads/main'
        run: tf-safe apply -auto-approve
```

#### GitLab CI
```yaml
terraform:
  image: hashicorp/terraform:latest
  before_script:
    - curl -fsSL https://raw.githubusercontent.com/your-org/tf-safe/main/scripts/install.sh | bash
    - tf-safe init --storage s3
  script:
    - tf-safe plan
    - tf-safe apply -auto-approve
  only:
    - main
```

## 🛠️ Building from Source

### Prerequisites
- Go 1.23 or later
- Git

### Build Steps
```bash
# Clone the repository
git clone https://github.com/your-org/tf-safe.git
cd tf-safe

# Build for your platform
make build

# Build for all platforms
make build-all

# Install locally
make install

# Run tests
make test
```

## 📚 Configuration Reference

### Local Storage Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable local backup storage |
| `path` | string | `.tfstate_snapshots` | Directory for local backups |
| `retention_count` | int | `10` | Number of local backups to retain |

### Remote Storage Options (S3)

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `provider` | string | `s3` | Storage provider (currently only s3) |
| `bucket` | string | `""` | S3 bucket name |
| `region` | string | `us-west-2` | AWS region |
| `prefix` | string | `""` | S3 key prefix |
| `enabled` | bool | `false` | Enable remote backup storage |

### Encryption Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `provider` | string | `aes` | Encryption provider (aes, kms, none) |
| `kms_key_id` | string | `""` | AWS KMS key ID (required for kms provider) |

### Retention Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `local_count` | int | `10` | Number of local backups to retain |
| `remote_count` | int | `50` | Number of remote backups to retain |
| `max_age_days` | int | `90` | Maximum backup age in days |

### Logging Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `level` | string | `info` | Log level (debug, info, warn, error) |
| `format` | string | `text` | Log format (text, json) |

## 🐛 Troubleshooting

### Common Issues

#### "terraform.tfstate not found"
tf-safe looks for `terraform.tfstate` in the current directory. Make sure you're running tf-safe from your Terraform project root.

#### "Permission denied" on backup directory
Ensure tf-safe has write permissions to the backup directory:
```bash
chmod 755 .tfstate_snapshots
```

#### AWS S3 authentication errors
Verify your AWS credentials are configured:
```bash
aws configure list
# or
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret
```

#### Backup corruption detected
If a backup fails integrity checks:
```bash
# List all backups to find a good one
tf-safe list

# Restore from a known good backup
tf-safe restore <backup-id>
```

### Debug Mode

Enable debug logging for troubleshooting:
```bash
tf-safe --log-level debug <command>
```

Or in configuration:
```yaml
logging:
  level: debug
```

### Getting Help

- **Documentation**: [GitHub Repository](https://github.com/your-org/tf-safe)
- **Issues**: [GitHub Issues](https://github.com/your-org/tf-safe/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/tf-safe/discussions)

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [HashiCorp Terraform](https://terraform.io) for the amazing infrastructure as code tool
- [Cobra](https://github.com/spf13/cobra) for the CLI framework
- [Viper](https://github.com/spf13/viper) for configuration management