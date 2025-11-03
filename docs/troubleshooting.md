# tf-safe Troubleshooting Guide

This guide helps you diagnose and resolve common issues with tf-safe.

## Quick Diagnostics

### Check tf-safe Status
```bash
# Verify installation
tf-safe --version

# Check configuration
tf-safe init --dry-run

# Test connectivity (if using remote storage)
tf-safe list --storage remote
```

### Enable Debug Logging
```bash
# Enable debug logging for any command
tf-safe --log-level debug <command>

# Or set in configuration
echo "logging:
  level: debug" >> .tf-safe.yaml
```

## Common Issues

### Installation Issues

#### "tf-safe: command not found"

**Symptoms:**
```bash
$ tf-safe --version
bash: tf-safe: command not found
```

**Solutions:**

1. **Verify installation location:**
   ```bash
   which tf-safe
   ls -la /usr/local/bin/tf-safe
   ```

2. **Check PATH:**
   ```bash
   echo $PATH
   # Ensure /usr/local/bin is in your PATH
   ```

3. **Reinstall using preferred method:**
   ```bash
   # Homebrew
   brew install tf-safe
   
   # Manual installation
   curl -fsSL https://raw.githubusercontent.com/your-org/tf-safe/main/scripts/install.sh | bash
   ```

4. **Add to PATH manually:**
   ```bash
   export PATH="/usr/local/bin:$PATH"
   echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
   ```

#### "Permission denied" when running tf-safe

**Symptoms:**
```bash
$ tf-safe --version
bash: /usr/local/bin/tf-safe: Permission denied
```

**Solutions:**

1. **Make binary executable:**
   ```bash
   sudo chmod +x /usr/local/bin/tf-safe
   ```

2. **Check file ownership:**
   ```bash
   ls -la /usr/local/bin/tf-safe
   sudo chown $(whoami) /usr/local/bin/tf-safe
   ```

### Configuration Issues

#### "Configuration file not found"

**Symptoms:**
```bash
$ tf-safe backup
Error: configuration file not found
```

**Solutions:**

1. **Initialize configuration:**
   ```bash
   tf-safe init
   ```

2. **Check configuration file location:**
   ```bash
   ls -la .tf-safe.yaml
   ls -la ~/.tf-safe/config.yaml
   ```

3. **Specify configuration file explicitly:**
   ```bash
   tf-safe --config /path/to/.tf-safe.yaml backup
   ```

#### "Invalid configuration" errors

**Symptoms:**
```bash
$ tf-safe backup
Error: Invalid configuration
- remote.bucket is required when remote.enabled=true
- retention.local_count must be at least 3 (got: 1)
```

**Solutions:**

1. **Fix configuration issues:**
   ```yaml
   # .tf-safe.yaml
   remote:
     enabled: true
     bucket: "my-terraform-backups"  # Add missing bucket
   
   retention:
     local_count: 5  # Increase from 1 to minimum 3
   ```

2. **Validate configuration:**
   ```bash
   tf-safe init --dry-run
   ```

3. **Use configuration template:**
   ```bash
   tf-safe init --force  # Recreate with defaults
   ```

### State File Issues

#### "terraform.tfstate not found"

**Symptoms:**
```bash
$ tf-safe backup
Error: terraform.tfstate not found in current directory
```

**Solutions:**

1. **Verify you're in the correct directory:**
   ```bash
   pwd
   ls -la terraform.tfstate
   ```

2. **Check for remote state:**
   ```bash
   # If using remote state, you might need to pull it first
   terraform state pull > terraform.tfstate
   ```

3. **Specify state file location:**
   ```yaml
   # .tf-safe.yaml
   terraform:
     state_file: "path/to/terraform.tfstate"
   ```

4. **Initialize Terraform if needed:**
   ```bash
   terraform init
   ```

#### "State file is empty or corrupted"

**Symptoms:**
```bash
$ tf-safe backup
Error: state file appears to be empty or corrupted
```

**Solutions:**

1. **Check state file content:**
   ```bash
   cat terraform.tfstate
   file terraform.tfstate
   ```

2. **Restore from backup if available:**
   ```bash
   tf-safe list
   tf-safe restore <backup-id>
   ```

3. **Refresh state from infrastructure:**
   ```bash
   terraform refresh
   ```

### Storage Issues

#### Local Storage Problems

**"Permission denied" creating backup directory:**

**Symptoms:**
```bash
$ tf-safe backup
Error: failed to create backup directory: permission denied
```

**Solutions:**

1. **Check directory permissions:**
   ```bash
   ls -la .tfstate_snapshots
   mkdir -p .tfstate_snapshots
   chmod 755 .tfstate_snapshots
   ```

2. **Change backup location:**
   ```yaml
   # .tf-safe.yaml
   local:
     path: "/tmp/tf-backups"  # Use writable directory
   ```

**"No space left on device":**

**Solutions:**

1. **Check disk space:**
   ```bash
   df -h .
   ```

2. **Clean up old backups:**
   ```bash
   tf-safe list
   # Manually remove old backups or reduce retention
   ```

3. **Reduce retention count:**
   ```yaml
   # .tf-safe.yaml
   retention:
     local_count: 3  # Reduce from higher number
   ```

#### Remote Storage (S3) Problems

**"Access denied" errors:**

**Symptoms:**
```bash
$ tf-safe backup
Error: failed to upload to S3: AccessDenied
```

**Solutions:**

1. **Check AWS credentials:**
   ```bash
   aws configure list
   aws sts get-caller-identity
   ```

2. **Verify S3 bucket permissions:**
   ```bash
   aws s3 ls s3://your-bucket-name/
   ```

3. **Check IAM policy:**
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "s3:GetObject",
           "s3:PutObject",
           "s3:DeleteObject",
           "s3:ListBucket"
         ],
         "Resource": [
           "arn:aws:s3:::your-bucket-name",
           "arn:aws:s3:::your-bucket-name/*"
         ]
       }
     ]
   }
   ```

4. **Test S3 access manually:**
   ```bash
   echo "test" | aws s3 cp - s3://your-bucket-name/test.txt
   aws s3 rm s3://your-bucket-name/test.txt
   ```

**"Bucket does not exist":**

**Solutions:**

1. **Create the bucket:**
   ```bash
   aws s3 mb s3://your-bucket-name --region us-west-2
   ```

2. **Verify bucket name and region:**
   ```bash
   aws s3 ls
   ```

**"Network timeout" errors:**

**Solutions:**

1. **Check network connectivity:**
   ```bash
   ping s3.amazonaws.com
   curl -I https://s3.amazonaws.com
   ```

2. **Increase timeout in configuration:**
   ```yaml
   # .tf-safe.yaml
   performance:
     retry_attempts: 5
     retry_delay: "10s"
   ```

3. **Use VPC endpoint if in AWS:**
   ```yaml
   # .tf-safe.yaml
   remote:
     s3:
       endpoint: "https://vpce-12345678-abcd-1234.s3.us-west-2.vpce.amazonaws.com"
   ```

### Encryption Issues

#### "Encryption key not found"

**Symptoms:**
```bash
$ tf-safe restore <backup-id>
Error: encryption key not found or invalid
```

**Solutions:**

1. **Check key file location:**
   ```bash
   ls -la ~/.tf-safe/keys/
   ```

2. **Regenerate key (will lose access to old backups):**
   ```bash
   rm ~/.tf-safe/keys/aes.key
   tf-safe backup  # Will generate new key
   ```

3. **Specify key file location:**
   ```yaml
   # .tf-safe.yaml
   encryption:
     aes:
       key_file: "/secure/path/tf-safe.key"
   ```

#### "KMS access denied"

**Symptoms:**
```bash
$ tf-safe backup
Error: KMS access denied for key arn:aws:kms:...
```

**Solutions:**

1. **Check KMS permissions:**
   ```bash
   aws kms describe-key --key-id arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012
   ```

2. **Verify KMS key policy:**
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Principal": {
           "AWS": "arn:aws:iam::123456789012:user/your-user"
         },
         "Action": [
           "kms:Encrypt",
           "kms:Decrypt",
           "kms:GenerateDataKey"
         ],
         "Resource": "*"
       }
     ]
   }
   ```

3. **Test KMS access:**
   ```bash
   echo "test" | aws kms encrypt --key-id arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012 --plaintext fileb://-
   ```

### Terraform Integration Issues

#### "Terraform binary not found"

**Symptoms:**
```bash
$ tf-safe apply
Error: terraform binary not found in PATH
```

**Solutions:**

1. **Install Terraform:**
   ```bash
   # Using tfenv
   tfenv install latest
   
   # Using Homebrew
   brew install terraform
   
   # Manual installation
   wget https://releases.hashicorp.com/terraform/1.6.0/terraform_1.6.0_linux_amd64.zip
   unzip terraform_1.6.0_linux_amd64.zip
   sudo mv terraform /usr/local/bin/
   ```

2. **Specify Terraform path:**
   ```yaml
   # .tf-safe.yaml
   terraform:
     binary_path: "/usr/local/bin/terraform"
   ```

3. **Check PATH:**
   ```bash
   which terraform
   echo $PATH
   ```

#### "Terraform command failed"

**Symptoms:**
```bash
$ tf-safe apply
Error: terraform command failed with exit code 1
```

**Solutions:**

1. **Run terraform command directly to see error:**
   ```bash
   terraform apply
   ```

2. **Check Terraform configuration:**
   ```bash
   terraform validate
   terraform plan
   ```

3. **Enable debug logging:**
   ```bash
   tf-safe --log-level debug apply
   ```

### Backup and Restore Issues

#### "Backup integrity check failed"

**Symptoms:**
```bash
$ tf-safe restore <backup-id>
Error: backup integrity check failed: checksum mismatch
```

**Solutions:**

1. **List available backups:**
   ```bash
   tf-safe list
   ```

2. **Try a different backup:**
   ```bash
   tf-safe restore <different-backup-id>
   ```

3. **Disable integrity check (not recommended):**
   ```yaml
   # .tf-safe.yaml
   verification:
     verify_on_restore: false
   ```

4. **Re-download from remote storage:**
   ```bash
   rm -rf .tfstate_snapshots
   tf-safe list --storage remote
   ```

#### "No backups found"

**Symptoms:**
```bash
$ tf-safe list
No backups found
```

**Solutions:**

1. **Check backup directories:**
   ```bash
   ls -la .tfstate_snapshots/
   ```

2. **Check remote storage:**
   ```bash
   tf-safe list --storage remote
   ```

3. **Create initial backup:**
   ```bash
   tf-safe backup
   ```

## Performance Issues

### Slow Backup Operations

**Solutions:**

1. **Increase concurrent uploads:**
   ```yaml
   # .tf-safe.yaml
   performance:
     concurrent_uploads: 5
     chunk_size: "128MB"
   ```

2. **Use compression:**
   ```yaml
   # .tf-safe.yaml
   compression:
     enabled: true
     algorithm: "gzip"
   ```

3. **Check network bandwidth:**
   ```bash
   speedtest-cli
   ```

### Large State Files

**Solutions:**

1. **Increase chunk size:**
   ```yaml
   # .tf-safe.yaml
   performance:
     chunk_size: "256MB"
   ```

2. **Use multipart uploads:**
   ```yaml
   # .tf-safe.yaml
   remote:
     s3:
       multipart_threshold: "100MB"
   ```

## Getting Help

### Collect Debug Information

When reporting issues, include:

1. **Version information:**
   ```bash
   tf-safe --version
   terraform --version
   ```

2. **Configuration (sanitized):**
   ```bash
   cat .tf-safe.yaml  # Remove sensitive information
   ```

3. **Debug logs:**
   ```bash
   tf-safe --log-level debug <failing-command> 2>&1 | tee debug.log
   ```

4. **Environment information:**
   ```bash
   uname -a
   echo $PATH
   env | grep -E "(AWS_|TF_)"
   ```

### Support Channels

- **GitHub Issues**: [Report bugs and feature requests](https://github.com/your-org/tf-safe/issues)
- **GitHub Discussions**: [Ask questions and share experiences](https://github.com/your-org/tf-safe/discussions)
- **Documentation**: [Complete documentation](https://github.com/your-org/tf-safe)

### Before Reporting Issues

1. **Search existing issues**: Check if your problem has been reported
2. **Try latest version**: Update to the latest tf-safe version
3. **Minimal reproduction**: Create a minimal example that reproduces the issue
4. **Include debug logs**: Always include debug output when reporting issues

## Emergency Recovery

### Complete State Loss

If you've lost your Terraform state completely:

1. **Check all backup locations:**
   ```bash
   tf-safe list --storage all
   ls -la .tfstate_snapshots/
   aws s3 ls s3://your-bucket/
   ```

2. **Import existing resources:**
   ```bash
   # If backups are unavailable, import resources manually
   terraform import aws_instance.example i-1234567890abcdef0
   ```

3. **Use terraform refresh:**
   ```bash
   # If you have partial state
   terraform refresh
   ```

### Corrupted Backups

If all backups appear corrupted:

1. **Check different storage backends:**
   ```bash
   tf-safe list --storage local
   tf-safe list --storage remote
   ```

2. **Try older backups:**
   ```bash
   tf-safe list | tail -10  # Try oldest backups
   ```

3. **Disable encryption temporarily:**
   ```yaml
   # .tf-safe.yaml
   encryption:
     provider: none
   ```

4. **Manual recovery from S3:**
   ```bash
   aws s3 cp s3://your-bucket/terraform.tfstate.2024-01-01T10:00:00Z.bak ./terraform.tfstate
   ```