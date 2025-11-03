# tf-safe Configuration Reference

This document provides a comprehensive reference for all tf-safe configuration options.

## Configuration Hierarchy

tf-safe uses a hierarchical configuration system where settings are merged in the following order (highest to lowest priority):

1. **Command-line flags** - Override all other settings
2. **Project-level configuration** - `.tf-safe.yaml` in project directory
3. **Global configuration** - `~/.tf-safe/config.yaml` in user home directory
4. **Built-in defaults** - Hardcoded fallback values

## Configuration File Locations

### Project-Level Configuration
```
/path/to/terraform/project/.tf-safe.yaml
```

### Global Configuration
```
# Linux/macOS
~/.tf-safe/config.yaml

# Windows
%USERPROFILE%\.tf-safe\config.yaml
```

## Complete Configuration Schema

```yaml
# Local storage backend configuration
local:
  enabled: true                    # Enable local backup storage
  path: ".tfstate_snapshots"      # Directory for storing local backups
  retention_count: 10             # Number of local backups to retain

# Remote storage backend configuration
remote:
  provider: "s3"                  # Storage provider (currently only "s3")
  bucket: ""                      # S3 bucket name (required if remote enabled)
  region: "us-west-2"            # AWS region
  prefix: ""                     # S3 key prefix (optional)
  enabled: false                 # Enable remote backup storage
  
  # S3-specific options
  s3:
    endpoint: ""                 # Custom S3 endpoint (for S3-compatible services)
    force_path_style: false      # Use path-style URLs instead of virtual-hosted
    server_side_encryption: ""   # Server-side encryption (AES256, aws:kms)
    sse_kms_key_id: ""          # KMS key ID for SSE-KMS

# Encryption configuration
encryption:
  provider: "aes"                # Encryption provider (aes, kms, none)
  kms_key_id: ""                # AWS KMS key ID (required for kms provider)
  
  # AES-specific options
  aes:
    key_file: ""               # Path to AES key file (auto-generated if empty)
    
  # Passphrase-based encryption
  passphrase:
    prompt: true               # Prompt for passphrase interactively
    env_var: "TF_SAFE_PASS"   # Environment variable containing passphrase

# Backup retention policies
retention:
  local_count: 10              # Number of local backups to retain
  remote_count: 50             # Number of remote backups to retain
  max_age_days: 90            # Maximum backup age in days (0 = no age limit)
  min_count: 3                # Minimum backups to always retain

# Terraform integration settings
terraform:
  binary_path: "terraform"     # Path to terraform binary
  auto_backup: true           # Enable automatic backups for wrapper commands
  backup_on_plan: true        # Create backup before terraform plan
  backup_on_apply: true       # Create backup before/after terraform apply
  backup_on_destroy: true     # Create backup before terraform destroy
  
  # Command timeout settings
  timeout: "30m"              # Maximum time for terraform commands
  
  # State file detection
  state_file: "terraform.tfstate"  # Expected state file name
  auto_detect: true           # Automatically detect state file location

# Logging configuration
logging:
  level: "info"               # Log level (debug, info, warn, error)
  format: "text"              # Log format (text, json)
  file: ""                    # Log file path (empty = stdout only)
  max_size: 100              # Maximum log file size in MB
  max_backups: 3             # Number of old log files to retain
  max_age: 28                # Maximum age of log files in days

# Backup verification settings
verification:
  checksum_algorithm: "sha256"  # Checksum algorithm (sha256, md5)
  verify_on_restore: true      # Verify backup integrity before restore
  verify_on_upload: true       # Verify backup after upload to remote storage

# Notification settings (future feature)
notifications:
  enabled: false
  webhook_url: ""
  slack_channel: ""
  email_recipients: []

# Performance tuning
performance:
  concurrent_uploads: 3        # Number of concurrent remote uploads
  chunk_size: "64MB"          # Upload chunk size for large files
  retry_attempts: 3           # Number of retry attempts for failed operations
  retry_delay: "5s"           # Initial delay between retries
```

## Configuration Sections

### Local Storage (`local`)

Controls local filesystem backup storage.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | boolean | `true` | Enable local backup storage |
| `path` | string | `.tfstate_snapshots` | Directory for storing local backups (relative to project root) |
| `retention_count` | integer | `10` | Number of local backups to retain (minimum 3) |

**Example:**
```yaml
local:
  enabled: true
  path: "backups/terraform"
  retention_count: 15
```

### Remote Storage (`remote`)

Controls remote cloud storage backup.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `provider` | string | `s3` | Storage provider (currently only "s3") |
| `bucket` | string | `""` | S3 bucket name (required if remote enabled) |
| `region` | string | `us-west-2` | AWS region |
| `prefix` | string | `""` | S3 key prefix for organizing backups |
| `enabled` | boolean | `false` | Enable remote backup storage |

**S3 Sub-options (`remote.s3`):**

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `endpoint` | string | `""` | Custom S3 endpoint for S3-compatible services |
| `force_path_style` | boolean | `false` | Use path-style URLs instead of virtual-hosted |
| `server_side_encryption` | string | `""` | Server-side encryption (AES256, aws:kms) |
| `sse_kms_key_id` | string | `""` | KMS key ID for SSE-KMS encryption |

**Example:**
```yaml
remote:
  provider: s3
  bucket: my-terraform-backups
  region: us-east-1
  prefix: "production/"
  enabled: true
  s3:
    server_side_encryption: "aws:kms"
    sse_kms_key_id: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
```

### Encryption (`encryption`)

Controls backup encryption settings.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `provider` | string | `aes` | Encryption provider (aes, kms, none) |
| `kms_key_id` | string | `""` | AWS KMS key ID (required for kms provider) |

**AES Sub-options (`encryption.aes`):**

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `key_file` | string | `""` | Path to AES key file (auto-generated if empty) |

**Passphrase Sub-options (`encryption.passphrase`):**

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `prompt` | boolean | `true` | Prompt for passphrase interactively |
| `env_var` | string | `TF_SAFE_PASS` | Environment variable containing passphrase |

**Examples:**

AES encryption:
```yaml
encryption:
  provider: aes
  aes:
    key_file: "/secure/path/tf-safe.key"
```

KMS encryption:
```yaml
encryption:
  provider: kms
  kms_key_id: "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012"
```

No encryption:
```yaml
encryption:
  provider: none
```

### Retention Policies (`retention`)

Controls backup retention and cleanup policies.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `local_count` | integer | `10` | Number of local backups to retain |
| `remote_count` | integer | `50` | Number of remote backups to retain |
| `max_age_days` | integer | `90` | Maximum backup age in days (0 = no limit) |
| `min_count` | integer | `3` | Minimum backups to always retain |

**Example:**
```yaml
retention:
  local_count: 5
  remote_count: 100
  max_age_days: 30
  min_count: 3
```

### Terraform Integration (`terraform`)

Controls Terraform command integration and behavior.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `binary_path` | string | `terraform` | Path to terraform binary |
| `auto_backup` | boolean | `true` | Enable automatic backups for wrapper commands |
| `backup_on_plan` | boolean | `true` | Create backup before terraform plan |
| `backup_on_apply` | boolean | `true` | Create backup before/after terraform apply |
| `backup_on_destroy` | boolean | `true` | Create backup before terraform destroy |
| `timeout` | duration | `30m` | Maximum time for terraform commands |
| `state_file` | string | `terraform.tfstate` | Expected state file name |
| `auto_detect` | boolean | `true` | Automatically detect state file location |

**Example:**
```yaml
terraform:
  binary_path: "/usr/local/bin/terraform"
  auto_backup: true
  backup_on_plan: false  # Skip backups for plan operations
  timeout: "45m"
```

### Logging (`logging`)

Controls logging behavior and output.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `level` | string | `info` | Log level (debug, info, warn, error) |
| `format` | string | `text` | Log format (text, json) |
| `file` | string | `""` | Log file path (empty = stdout only) |
| `max_size` | integer | `100` | Maximum log file size in MB |
| `max_backups` | integer | `3` | Number of old log files to retain |
| `max_age` | integer | `28` | Maximum age of log files in days |

**Example:**
```yaml
logging:
  level: debug
  format: json
  file: "/var/log/tf-safe.log"
  max_size: 50
```

## Environment Variables

tf-safe supports configuration via environment variables. Environment variables use the prefix `TF_SAFE_` followed by the configuration path in uppercase with underscores.

### Common Environment Variables

| Variable | Configuration Path | Description |
|----------|-------------------|-------------|
| `TF_SAFE_LOCAL_ENABLED` | `local.enabled` | Enable local storage |
| `TF_SAFE_REMOTE_BUCKET` | `remote.bucket` | S3 bucket name |
| `TF_SAFE_REMOTE_REGION` | `remote.region` | AWS region |
| `TF_SAFE_ENCRYPTION_PROVIDER` | `encryption.provider` | Encryption provider |
| `TF_SAFE_ENCRYPTION_KMS_KEY_ID` | `encryption.kms_key_id` | KMS key ID |
| `TF_SAFE_LOGGING_LEVEL` | `logging.level` | Log level |

### AWS Credentials

tf-safe uses standard AWS credential resolution:

1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. Shared credentials file (`~/.aws/credentials`)
3. IAM roles (for EC2 instances)
4. AWS profiles (`AWS_PROFILE`)

## Command-Line Flags

Global flags that override configuration file settings:

| Flag | Configuration Path | Description |
|------|-------------------|-------------|
| `--config` | - | Path to configuration file |
| `--log-level` | `logging.level` | Log level |
| `--log-format` | `logging.format` | Log format |
| `--no-encrypt` | `encryption.provider=none` | Disable encryption |
| `--local-only` | `remote.enabled=false` | Use only local storage |
| `--remote-only` | `local.enabled=false` | Use only remote storage |

## Configuration Validation

tf-safe validates configuration on startup and provides helpful error messages:

### Required Fields
- `remote.bucket` (if `remote.enabled=true`)
- `encryption.kms_key_id` (if `encryption.provider=kms`)

### Validation Rules
- `retention.local_count` must be ≥ 3
- `retention.remote_count` must be ≥ 1
- `retention.max_age_days` must be ≥ 0
- `logging.level` must be one of: debug, info, warn, error
- `encryption.provider` must be one of: aes, kms, none

### Example Validation Errors

```
Error: Invalid configuration
- remote.bucket is required when remote.enabled=true
- retention.local_count must be at least 3 (got: 1)
- encryption.kms_key_id is required when encryption.provider=kms
```

## Configuration Examples

### Minimal Local-Only Setup
```yaml
local:
  enabled: true
  retention_count: 5

encryption:
  provider: aes
```

### Production Setup with S3 and KMS
```yaml
local:
  enabled: true
  retention_count: 10

remote:
  provider: s3
  bucket: prod-terraform-backups
  region: us-east-1
  prefix: "infrastructure/"
  enabled: true

encryption:
  provider: kms
  kms_key_id: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"

retention:
  local_count: 10
  remote_count: 100
  max_age_days: 365

logging:
  level: info
  format: json
```

### Development Setup
```yaml
local:
  enabled: true
  path: ".tf-backups"
  retention_count: 3

encryption:
  provider: none

terraform:
  backup_on_plan: false

logging:
  level: debug
```

### CI/CD Setup
```yaml
local:
  enabled: false

remote:
  provider: s3
  bucket: ci-terraform-backups
  region: us-west-2
  prefix: "${CI_PROJECT_NAME}/"
  enabled: true

encryption:
  provider: kms
  kms_key_id: "${KMS_KEY_ID}"

retention:
  remote_count: 50
  max_age_days: 90

logging:
  level: info
  format: json
```