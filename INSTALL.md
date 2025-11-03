# tf-safe Installation Guide

tf-safe is available for Linux, macOS, and Windows. Choose the installation method that works best for your environment.

## Quick Install

### macOS (Homebrew)

```bash
brew install tf-safe
```

### Linux (APT - Debian/Ubuntu)

```bash
# Add repository (if available)
curl -fsSL https://packages.tf-safe.dev/gpg | sudo apt-key add -
echo "deb https://packages.tf-safe.dev/apt stable main" | sudo tee /etc/apt/sources.list.d/tf-safe.list

# Install
sudo apt update
sudo apt install tf-safe
```

### Windows (Chocolatey)

```powershell
choco install tf-safe
```

## Manual Installation

### Download Pre-built Binaries

1. Go to the [releases page](https://github.com/BIRhrt/tf-safe/releases)
2. Download the appropriate binary for your platform:
   - **Linux AMD64**: `tf-safe-linux-amd64.tar.gz`
   - **Linux ARM64**: `tf-safe-linux-arm64.tar.gz`
   - **macOS AMD64**: `tf-safe-darwin-amd64.tar.gz`
   - **macOS ARM64**: `tf-safe-darwin-arm64.tar.gz`
   - **Windows AMD64**: `tf-safe-windows-amd64.zip`

### Linux/macOS Installation

```bash
# Download and extract (replace with your platform)
curl -LO https://github.com/BIRhrt/tf-safe/releases/latest/download/tf-safe-linux-amd64.tar.gz
tar -xzf tf-safe-linux-amd64.tar.gz

# Make executable and move to PATH
chmod +x tf-safe-linux-amd64
sudo mv tf-safe-linux-amd64 /usr/local/bin/tf-safe

# Verify installation
tf-safe --version
```

### Windows Installation

1. Download `tf-safe-windows-amd64.zip`
2. Extract the ZIP file
3. Move `tf-safe.exe` to a directory in your PATH
4. Open Command Prompt or PowerShell and run `tf-safe --version`

## Installation Script

For Linux and macOS, you can use our installation script:

```bash
curl -fsSL https://raw.githubusercontent.com/BIRhrt/tf-safe/main/scripts/install.sh | bash
```

Or download and inspect first:

```bash
curl -fsSL https://raw.githubusercontent.com/BIRhrt/tf-safe/main/scripts/install.sh -o install.sh
chmod +x install.sh
./install.sh
```

## Build from Source

### Prerequisites

- Go 1.23 or later
- Git

### Build Steps

```bash
# Clone the repository
git clone https://github.com/BIRhrt/tf-safe.git
cd tf-safe

# Build for your platform
make build

# Or build for all platforms
make build-all

# Install locally
make install
```

## Verify Installation

After installation, verify tf-safe is working:

```bash
# Check version
tf-safe --version

# View help
tf-safe --help

# Initialize a project (creates .tf-safe.yaml)
cd /path/to/terraform/project
tf-safe init
```

## Shell Completion

tf-safe supports shell completion for bash, zsh, fish, and PowerShell.

### Bash

```bash
# Add to ~/.bashrc
echo 'source <(tf-safe completion bash)' >> ~/.bashrc
source ~/.bashrc
```

### Zsh

```bash
# Add to ~/.zshrc
echo 'source <(tf-safe completion zsh)' >> ~/.zshrc
source ~/.zshrc
```

### Fish

```bash
tf-safe completion fish | source
```

### PowerShell

```powershell
tf-safe completion powershell | Out-String | Invoke-Expression
```

## Uninstallation

### Package Managers

```bash
# Homebrew
brew uninstall tf-safe

# APT
sudo apt remove tf-safe

# Chocolatey
choco uninstall tf-safe
```

### Manual Installation

```bash
# Remove binary
sudo rm /usr/local/bin/tf-safe

# Remove configuration (optional)
rm -rf ~/.tf-safe
```

## Troubleshooting

### Permission Denied

If you get permission denied errors:

```bash
# Make sure the binary is executable
chmod +x /usr/local/bin/tf-safe

# Check if /usr/local/bin is in your PATH
echo $PATH
```

### Command Not Found

If `tf-safe` command is not found:

1. Verify the binary is in your PATH
2. Restart your terminal
3. Check installation location: `which tf-safe`

### Version Mismatch

If you're seeing an old version:

```bash
# Check which binary is being used
which tf-safe

# Remove old installations
sudo rm /usr/bin/tf-safe /usr/local/bin/tf-safe

# Reinstall using your preferred method
```

## Getting Help

- **Documentation**: [GitHub Repository](https://github.com/BIRhrt/tf-safe)
- **Issues**: [GitHub Issues](https://github.com/BIRhrt/tf-safe/issues)
- **Discussions**: [GitHub Discussions](https://github.com/BIRhrt/tf-safe/discussions)

## Next Steps

After installation:

1. **Initialize a project**: `tf-safe init`
2. **Configure backends**: Edit `.tf-safe.yaml`
3. **Start using**: Replace `terraform` with `tf-safe` in your commands
4. **View backups**: `tf-safe list`

For detailed usage instructions, see the [README](README.md).