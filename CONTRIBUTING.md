# Contributing to tf-safe

Thank you for your interest in contributing to tf-safe! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites

- Go 1.23 or later
- Git
- Make (optional, but recommended)

### Development Setup

1. **Fork and clone the repository**:
   ```bash
   git clone https://github.com/your-username/tf-safe.git
   cd tf-safe
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Build the project**:
   ```bash
   make build
   # or
   go build -o tf-safe .
   ```

4. **Run tests**:
   ```bash
   make test
   # or
   go test ./...
   ```

## 📝 How to Contribute

### Reporting Issues

Before creating an issue, please:

1. **Search existing issues** to avoid duplicates
2. **Use the issue templates** when available
3. **Provide detailed information**:
   - tf-safe version (`tf-safe --version`)
   - Operating system and version
   - Go version (`go version`)
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant logs (use `--log-level debug`)

### Suggesting Features

Feature requests are welcome! Please:

1. **Check existing feature requests** first
2. **Describe the use case** clearly
3. **Explain why** this feature would be valuable
4. **Consider the scope** - start with smaller, focused features

### Code Contributions

#### Pull Request Process

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:
   - Follow the coding standards (see below)
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**:
   ```bash
   make test
   make lint
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add new backup encryption method"
   ```

5. **Push and create a pull request**:
   ```bash
   git push origin feature/your-feature-name
   ```

#### Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(backup): add support for Azure Blob Storage
fix(config): resolve validation error for empty bucket name
docs: update installation instructions for Windows
test(storage): add integration tests for S3 backend
```

## 🏗️ Development Guidelines

### Code Style

- **Follow Go conventions**: Use `gofmt`, `golint`, and `go vet`
- **Use meaningful names**: Variables, functions, and packages should have descriptive names
- **Keep functions small**: Aim for single responsibility
- **Add comments**: Especially for exported functions and complex logic
- **Handle errors properly**: Don't ignore errors, wrap them with context

### Project Structure

```
tf-safe/
├── cmd/                    # CLI commands (Cobra)
├── internal/               # Private application code
│   ├── backup/            # Backup engine
│   ├── config/            # Configuration management
│   ├── encryption/        # Encryption utilities
│   ├── restore/           # Restoration engine
│   ├── storage/           # Storage backends
│   └── terraform/         # Terraform integration
├── pkg/                   # Public library code
│   └── types/             # Shared types
├── test/                  # Test files
├── docs/                  # Documentation
├── examples/              # Usage examples
└── scripts/               # Build and deployment scripts
```

### Testing

- **Write tests** for all new functionality
- **Use table-driven tests** where appropriate
- **Mock external dependencies** (AWS, filesystem, etc.)
- **Test error conditions** and edge cases
- **Maintain test coverage** above 80%

**Test Types:**
- **Unit tests**: Test individual functions/methods
- **Integration tests**: Test component interactions
- **End-to-end tests**: Test complete workflows

**Running Tests:**
```bash
# All tests
make test

# Unit tests only
go test ./internal/...

# Integration tests
go test ./test/integration/...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

### Adding New Storage Backends

To add a new storage backend:

1. **Implement the interface**:
   ```go
   // internal/storage/interface.go
   type Backend interface {
       Store(ctx context.Context, key string, data []byte) error
       Retrieve(ctx context.Context, key string) ([]byte, error)
       List(ctx context.Context, prefix string) ([]string, error)
       Delete(ctx context.Context, key string) error
   }
   ```

2. **Create the implementation**:
   ```go
   // internal/storage/newbackend.go
   type NewBackend struct {
       // configuration fields
   }

   func (b *NewBackend) Store(ctx context.Context, key string, data []byte) error {
       // implementation
   }
   // ... other methods
   ```

3. **Add configuration support**:
   ```go
   // pkg/types/config.go
   type RemoteConfig struct {
       Provider string `yaml:"provider"`
       // ... existing fields
       NewBackend *NewBackendConfig `yaml:"new_backend,omitempty"`
   }
   ```

4. **Register in factory**:
   ```go
   // internal/storage/factory.go
   func NewBackend(config *types.RemoteConfig) (Backend, error) {
       switch config.Provider {
       case "s3":
           return NewS3Backend(config.S3)
       case "new_backend":
           return NewNewBackend(config.NewBackend)
       // ...
       }
   }
   ```

5. **Add tests and documentation**

### Adding New Encryption Methods

Similar process for encryption providers:

1. **Implement the interface** in `internal/encryption/`
2. **Add configuration** in `pkg/types/config.go`
3. **Register in factory** in `internal/encryption/factory.go`
4. **Add comprehensive tests**

## 🧪 Testing Guidelines

### Unit Tests

```go
func TestBackupEngine_CreateBackup(t *testing.T) {
    tests := []struct {
        name    string
        setup   func() *BackupEngine
        want    error
        wantErr bool
    }{
        {
            name: "successful backup",
            setup: func() *BackupEngine {
                // setup test case
            },
            wantErr: false,
        },
        // more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            engine := tt.setup()
            err := engine.CreateBackup(context.Background())
            
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateBackup() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

```go
func TestS3Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup real S3 connection or localstack
    backend := setupS3Backend(t)
    defer cleanupS3Backend(t, backend)

    // Test real operations
    err := backend.Store(context.Background(), "test-key", []byte("test-data"))
    require.NoError(t, err)
}
```

## 📚 Documentation

### Code Documentation

- **Document all exported functions** with Go doc comments
- **Include examples** in doc comments where helpful
- **Document complex algorithms** with inline comments

### User Documentation

When adding features that affect users:

1. **Update README.md** if it changes core functionality
2. **Update docs/configuration.md** for new configuration options
3. **Add examples** in the `examples/` directory
4. **Update INSTALL.md** for installation changes

## 🔍 Code Review Process

### For Contributors

- **Keep PRs focused** - one feature/fix per PR
- **Write clear descriptions** of what changed and why
- **Respond to feedback** promptly and constructively
- **Update your PR** based on review comments

### For Reviewers

- **Be constructive** and helpful in feedback
- **Focus on code quality**, not personal preferences
- **Test the changes** locally when possible
- **Approve when ready** - don't let perfect be the enemy of good

## 🏷️ Release Process

Releases are handled by maintainers:

1. **Version bump** in appropriate files
2. **Update CHANGELOG.md** with release notes
3. **Create and push tag**: `git tag v1.2.3 && git push origin v1.2.3`
4. **GitHub Actions** automatically builds and publishes release
5. **Update package managers** (Homebrew, Chocolatey, etc.)

## 🤝 Community

### Getting Help

- **GitHub Discussions**: Ask questions and share ideas
- **GitHub Issues**: Report bugs and request features
- **Code Reviews**: Learn from feedback on your PRs

### Code of Conduct

We follow the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). Please be respectful and inclusive in all interactions.

## 📄 License

By contributing to tf-safe, you agree that your contributions will be licensed under the MIT License.

## 🙏 Recognition

Contributors are recognized in:
- **GitHub contributors page**
- **Release notes** for significant contributions
- **README.md** for major features

Thank you for contributing to tf-safe! 🎉