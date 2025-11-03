# Simple AWS Infrastructure with tf-safe

This example demonstrates how to use tf-safe with a basic AWS infrastructure setup.

## What This Example Creates

- VPC with public subnet
- Internet Gateway and routing
- Security group allowing HTTP/HTTPS traffic
- EC2 instance running a simple web server

## Prerequisites

1. **AWS CLI configured** with appropriate credentials
2. **Terraform installed** (version 1.0+)
3. **tf-safe installed** and configured

## Setup Instructions

### 1. Configure AWS Credentials

```bash
aws configure
# or
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_DEFAULT_REGION=us-west-2
```

### 2. Initialize tf-safe

```bash
cd examples/terraform-projects/simple-aws
tf-safe init
```

This will create a `.tf-safe.yaml` configuration file. Edit it to match your preferences:

```yaml
# Enable S3 remote storage (optional)
remote:
  provider: s3
  bucket: "your-terraform-backups"
  region: "us-west-2"
  prefix: "simple-aws-demo/"
  enabled: true
```

### 3. Initialize Terraform

```bash
tf-safe init  # This runs terraform init with backup hooks
# or
terraform init
```

### 4. Plan and Apply

```bash
# Plan with automatic backup
tf-safe plan

# Apply with automatic backups
tf-safe apply

# Check the web server
curl http://$(terraform output -raw web_instance_public_ip)
```

## tf-safe Integration Points

### Automatic Backups

tf-safe automatically creates backups:

- **Before terraform apply**: Protects current state
- **After terraform apply**: Captures new state
- **Before terraform destroy**: Preserves state before deletion

### Manual Backup Operations

```bash
# Create manual backup
tf-safe backup --message "Before major changes"

# List all backups
tf-safe list

# Restore a specific backup
tf-safe restore 2024-01-15T10:30:00Z
```

### Backup Locations

With the default configuration:

- **Local backups**: `.tfstate_snapshots/` directory
- **Remote backups**: S3 bucket (if configured)

## Workflow Examples

### Development Workflow

```bash
# 1. Make infrastructure changes
vim main.tf

# 2. Plan changes (creates backup)
tf-safe plan

# 3. Apply changes (creates before/after backups)
tf-safe apply

# 4. If something goes wrong, restore previous state
tf-safe list
tf-safe restore <backup-id>
```

### Testing Disaster Recovery

```bash
# 1. Create a known good backup
tf-safe backup --message "Known good state"

# 2. Make some changes
tf-safe apply

# 3. Simulate state loss
rm terraform.tfstate

# 4. Restore from backup
tf-safe restore <backup-id>

# 5. Verify infrastructure is intact
terraform plan  # Should show no changes
```

## Configuration Customization

### Local Development

```yaml
# .tf-safe.yaml
local:
  enabled: true
  retention_count: 5

encryption:
  provider: none  # Disable encryption for simplicity

terraform:
  backup_on_plan: false  # Skip plan backups
```

### Production Setup

```yaml
# .tf-safe.yaml
local:
  enabled: true
  retention_count: 10

remote:
  provider: s3
  bucket: "prod-terraform-backups"
  enabled: true

encryption:
  provider: kms
  kms_key_id: "arn:aws:kms:us-west-2:123456789012:key/..."

retention:
  local_count: 10
  remote_count: 100
  max_age_days: 365
```

## Cleanup

To destroy the infrastructure:

```bash
# Destroy with automatic backup
tf-safe destroy

# Or use terraform directly
terraform destroy
```

The backup created before destruction allows you to restore the infrastructure later if needed.

## Troubleshooting

### Common Issues

1. **AWS credentials not configured**:
   ```bash
   aws configure list
   ```

2. **S3 bucket doesn't exist**:
   ```bash
   aws s3 mb s3://your-terraform-backups
   ```

3. **Permission denied on backup directory**:
   ```bash
   chmod 755 .tfstate_snapshots
   ```

### Debug Mode

Enable debug logging to troubleshoot issues:

```bash
tf-safe --log-level debug apply
```

## Next Steps

- Try the [multi-environment example](../multi-environment/) for more complex scenarios
- Explore [CI/CD integration examples](../../ci-cd/) for automated workflows
- Review the [configuration reference](../../docs/configuration.md) for advanced options