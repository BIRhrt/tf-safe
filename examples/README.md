# tf-safe Examples

This directory contains comprehensive examples demonstrating how to use tf-safe in various scenarios and environments.

## 📁 Directory Structure

```
examples/
├── configurations/          # Example configuration files
│   ├── minimal-local.yaml          # Basic local-only setup
│   ├── production-s3-kms.yaml      # Production with S3 and KMS
│   ├── development.yaml            # Development environment
│   ├── team-shared.yaml            # Team collaboration setup
│   ├── ci-cd.yaml                  # CI/CD pipeline configuration
│   └── multi-environment.yaml      # Multi-environment setup
├── terraform-projects/      # Sample Terraform projects
│   └── simple-aws/                 # Basic AWS infrastructure
├── ci-cd/                   # CI/CD pipeline examples
│   ├── github-actions.yml          # GitHub Actions workflow
│   ├── gitlab-ci.yml               # GitLab CI/CD pipeline
│   └── azure-devops.yml            # Azure DevOps pipeline
└── README.md               # This file
```

## 🚀 Quick Start

### 1. Choose a Configuration

Start with one of the example configurations based on your needs:

- **Individual Developer**: Use `configurations/minimal-local.yaml`
- **Team Environment**: Use `configurations/team-shared.yaml`
- **Production Setup**: Use `configurations/production-s3-kms.yaml`
- **CI/CD Pipeline**: Use `configurations/ci-cd.yaml`

### 2. Copy and Customize

```bash
# Copy an example configuration
cp examples/configurations/minimal-local.yaml .tf-safe.yaml

# Edit to match your environment
vim .tf-safe.yaml
```

### 3. Initialize and Test

```bash
# Initialize tf-safe
tf-safe init

# Test with a sample project
cd examples/terraform-projects/simple-aws
tf-safe plan
```

## 📋 Configuration Examples

### Minimal Local Setup

Perfect for individual developers or small projects:

```yaml
local:
  enabled: true
  retention_count: 5

encryption:
  provider: aes

terraform:
  backup_on_plan: false  # Save space
```

**Use case**: Local development, learning, small personal projects

### Production Setup

Enterprise-grade configuration with full security:

```yaml
local:
  enabled: true
  retention_count: 10

remote:
  provider: s3
  bucket: "prod-terraform-backups"
  enabled: true

encryption:
  provider: kms
  kms_key_id: "arn:aws:kms:..."

retention:
  remote_count: 100
  max_age_days: 365
```

**Use case**: Production environments, compliance requirements, enterprise teams

### Team Collaboration

Shared configuration for team environments:

```yaml
remote:
  provider: s3
  bucket: "team-terraform-backups"
  prefix: "${PROJECT_NAME}/"
  enabled: true

encryption:
  provider: aes
  aes:
    key_file: "/shared/secure/tf-safe-team.key"
```

**Use case**: Team projects, shared infrastructure, collaborative development

### CI/CD Pipeline

Optimized for automated environments:

```yaml
local:
  enabled: false  # No local storage in CI/CD

remote:
  provider: s3
  bucket: "ci-terraform-backups"
  prefix: "${CI_PROJECT_NAME}/${CI_ENVIRONMENT_NAME}/"
  enabled: true

logging:
  format: json  # Structured logging
```

**Use case**: GitHub Actions, GitLab CI, Azure DevOps, Jenkins

## 🏗️ Terraform Project Examples

### Simple AWS Infrastructure

The `terraform-projects/simple-aws/` example demonstrates:

- Basic AWS VPC setup
- EC2 instance with web server
- tf-safe integration
- Development workflow

**Getting Started:**

```bash
cd examples/terraform-projects/simple-aws
tf-safe init
tf-safe apply
```

**What it creates:**
- VPC with public subnet
- Internet Gateway and routing
- Security group for web traffic
- EC2 instance with simple web server

**tf-safe features demonstrated:**
- Automatic backups before/after apply
- Local backup storage
- Configuration management
- State restoration

## 🔄 CI/CD Integration Examples

### GitHub Actions

Complete workflow with:
- Pull request validation
- Automatic deployments
- Backup management
- Failure notifications

**Key features:**
- Environment-specific configurations
- Automatic backup creation
- Rollback instructions on failure
- Security best practices

### GitLab CI/CD

Multi-environment pipeline with:
- Development, staging, production stages
- Manual approval gates
- Emergency restore jobs
- Comprehensive logging

**Key features:**
- Environment isolation
- Backup verification
- Scheduled cleanup
- Manual destroy operations

### Azure DevOps

Enterprise pipeline with:
- Multi-stage deployments
- Artifact management
- Environment approvals
- Restore pipeline

**Key features:**
- Variable groups for secrets
- Environment-specific deployments
- Backup artifact publishing
- Emergency restore procedures

## 🛠️ Customization Guide

### Environment Variables

Use environment variables for dynamic configuration:

```yaml
remote:
  bucket: "${TF_BACKUP_BUCKET}"
  prefix: "${PROJECT_NAME}/${ENVIRONMENT}/"

encryption:
  kms_key_id: "${KMS_KEY_ID}"
```

**Common variables:**
- `TF_BACKUP_BUCKET`: S3 bucket name
- `KMS_KEY_ID`: AWS KMS key ID
- `PROJECT_NAME`: Project identifier
- `ENVIRONMENT`: Environment name (dev, staging, prod)

### Multi-Environment Setup

For organizations managing multiple environments:

```yaml
# Base configuration
retention:
  remote_count: 50
  max_age_days: 90

# Environment-specific overrides via variables
# Development: TF_SAFE_RETENTION_REMOTE_COUNT=20
# Production: TF_SAFE_RETENTION_REMOTE_COUNT=200
```

### Security Considerations

#### Development Environment
```yaml
encryption:
  provider: none  # Faster, less secure

logging:
  level: debug    # Detailed logging
```

#### Production Environment
```yaml
encryption:
  provider: kms
  kms_key_id: "arn:aws:kms:..."

verification:
  verify_on_restore: true
  verify_on_upload: true

logging:
  level: info
  format: json
```

## 📊 Monitoring and Alerting

### Backup Monitoring

Monitor backup operations in your CI/CD:

```bash
# Check backup count
BACKUP_COUNT=$(tf-safe list | wc -l)
if [ $BACKUP_COUNT -lt 5 ]; then
  echo "WARNING: Low backup count: $BACKUP_COUNT"
fi

# Check recent backup age
LATEST_BACKUP=$(tf-safe list --limit 1 --format json | jq -r '.[0].timestamp')
# Add age checking logic
```

### Failure Notifications

Example Slack notification on failure:

```bash
if [ $? -ne 0 ]; then
  curl -X POST -H 'Content-type: application/json' \
    --data '{"text":"Terraform apply failed. Backups available for rollback."}' \
    $SLACK_WEBHOOK_URL
fi
```

## 🔧 Troubleshooting

### Common Issues

1. **S3 Access Denied**
   ```bash
   # Check AWS credentials
   aws sts get-caller-identity
   
   # Test S3 access
   aws s3 ls s3://your-bucket/
   ```

2. **KMS Permission Errors**
   ```bash
   # Test KMS access
   aws kms describe-key --key-id your-key-id
   ```

3. **Configuration Validation**
   ```bash
   # Validate configuration
   tf-safe init --dry-run
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

## 📚 Additional Resources

- [Configuration Reference](../docs/configuration.md)
- [Troubleshooting Guide](../docs/troubleshooting.md)
- [Installation Guide](../INSTALL.md)
- [Main Documentation](../README.md)

## 🤝 Contributing

Found an issue with an example or want to add a new one?

1. Check existing [issues](https://github.com/BIRhrt/tf-safe/issues)
2. Create a new issue or pull request
3. Follow the [contribution guidelines](../CONTRIBUTING.md)

## 📄 License

These examples are provided under the same license as tf-safe. See [LICENSE](../LICENSE) for details.