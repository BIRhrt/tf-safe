package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"tf-safe/internal/utils"
	tftypes "tf-safe/pkg/types"
)

const (
	// S3MetadataPrefix is the prefix for S3 object metadata
	S3MetadataPrefix = "tf-safe-"
	// S3MultipartThreshold is the size threshold for multipart uploads (5MB)
	S3MultipartThreshold = 5 * 1024 * 1024
	// S3MaxRetries is the maximum number of retry attempts
	S3MaxRetries = 3
	// S3RetryDelay is the base delay for exponential backoff
	S3RetryDelay = time.Second
)

// S3Storage implements StorageBackend for AWS S3 storage
type S3Storage struct {
	config tftypes.RemoteConfig
	client *s3.Client
	logger *utils.Logger
}

// NewS3Storage creates a new S3 storage backend
func NewS3Storage(remoteConfig tftypes.RemoteConfig, logger *utils.Logger) *S3Storage {
	return &S3Storage{
		config: remoteConfig,
		logger: logger,
	}
}

// Initialize sets up the S3 storage backend
func (s3s *S3Storage) Initialize(ctx context.Context) error {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(s3s.config.Region))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	s3s.client = s3.NewFromConfig(cfg)

	// Validate S3 connectivity and permissions
	if err := s3s.validateS3Access(ctx); err != nil {
		return fmt.Errorf("S3 validation failed: %w", err)
	}

	s3s.logger.Info("S3 storage initialized for bucket %s in region %s", 
		s3s.config.Bucket, s3s.config.Region)
	return nil
}

// Store saves backup data to S3
func (s3s *S3Storage) Store(ctx context.Context, key string, data []byte, metadata *tftypes.BackupMetadata) error {
	s3Key := s3s.buildS3Key(key)
	
	// Calculate checksum if not provided
	if metadata.Checksum == "" {
		metadata.Checksum = utils.CalculateChecksumBytes(data)
	}

	// Update metadata
	metadata.Size = int64(len(data))
	metadata.StorageType = s3s.GetType()
	metadata.FilePath = fmt.Sprintf("s3://%s/%s", s3s.config.Bucket, s3Key)

	// Prepare S3 metadata
	s3Metadata := map[string]string{
		S3MetadataPrefix + "id":          metadata.ID,
		S3MetadataPrefix + "timestamp":   metadata.Timestamp.Format(time.RFC3339),
		S3MetadataPrefix + "checksum":    metadata.Checksum,
		S3MetadataPrefix + "encrypted":   fmt.Sprintf("%t", metadata.Encrypted),
		S3MetadataPrefix + "size":        fmt.Sprintf("%d", metadata.Size),
	}

	// Use multipart upload for large files
	if len(data) > S3MultipartThreshold {
		return s3s.multipartUpload(ctx, s3Key, data, s3Metadata)
	}

	// Regular upload for smaller files
	return s3s.regularUpload(ctx, s3Key, data, s3Metadata)
}

// Retrieve gets backup data from S3
func (s3s *S3Storage) Retrieve(ctx context.Context, key string) ([]byte, *tftypes.BackupMetadata, error) {
	s3Key := s3s.buildS3Key(key)

	// Get object with retry logic
	var getOutput *s3.GetObjectOutput
	var err error
	
	for attempt := 0; attempt < S3MaxRetries; attempt++ {
		getOutput, err = s3s.client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s3s.config.Bucket),
			Key:    aws.String(s3Key),
		})
		
		if err == nil {
			break
		}
		
		if attempt < S3MaxRetries-1 {
			delay := time.Duration(attempt+1) * S3RetryDelay
			s3s.logger.Warn("S3 GetObject attempt %d failed, retrying in %v: %v", 
				attempt+1, delay, err)
			time.Sleep(delay)
		}
	}
	
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve object from S3 after %d attempts: %w", 
			S3MaxRetries, err)
	}
	defer getOutput.Body.Close()

	// Read object data
	data, err := io.ReadAll(getOutput.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read S3 object data: %w", err)
	}

	// Parse metadata from S3 object metadata
	metadata, err := s3s.parseS3Metadata(getOutput.Metadata, key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse S3 metadata: %w", err)
	}

	// Validate checksum
	actualChecksum := utils.CalculateChecksumBytes(data)
	if actualChecksum != metadata.Checksum {
		return nil, nil, fmt.Errorf("checksum mismatch for backup %s: expected %s, got %s", 
			key, metadata.Checksum, actualChecksum)
	}

	s3s.logger.Debug("Backup retrieved successfully from S3: %s", key)
	return data, metadata, nil
}

// List returns all available backups in S3
func (s3s *S3Storage) List(ctx context.Context) ([]*tftypes.BackupMetadata, error) {
	var backups []*tftypes.BackupMetadata
	prefix := s3s.config.Prefix

	// List objects with retry logic
	var listOutput *s3.ListObjectsV2Output
	var err error
	
	for attempt := 0; attempt < S3MaxRetries; attempt++ {
		listOutput, err = s3s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(s3s.config.Bucket),
			Prefix: aws.String(prefix),
		})
		
		if err == nil {
			break
		}
		
		if attempt < S3MaxRetries-1 {
			delay := time.Duration(attempt+1) * S3RetryDelay
			s3s.logger.Warn("S3 ListObjectsV2 attempt %d failed, retrying in %v: %v", 
				attempt+1, delay, err)
			time.Sleep(delay)
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects after %d attempts: %w", 
			S3MaxRetries, err)
	}

	// Process each object
	for _, obj := range listOutput.Contents {
		if obj.Key == nil || !strings.HasSuffix(*obj.Key, BackupFileExtension) {
			continue
		}

		// Extract backup key from S3 key
		backupKey := s3s.extractBackupKey(*obj.Key)
		if backupKey == "" {
			continue
		}

		// Get object metadata
		headOutput, err := s3s.client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(s3s.config.Bucket),
			Key:    obj.Key,
		})
		if err != nil {
			s3s.logger.Warn("Failed to get metadata for S3 object %s: %v", *obj.Key, err)
			continue
		}

		// Parse metadata
		metadata, err := s3s.parseS3Metadata(headOutput.Metadata, backupKey)
		if err != nil {
			s3s.logger.Warn("Failed to parse metadata for S3 object %s: %v", *obj.Key, err)
			continue
		}

		// Update metadata with S3-specific information
		if obj.Size != nil {
			metadata.Size = *obj.Size
		}
		metadata.StorageType = s3s.GetType()
		metadata.FilePath = fmt.Sprintf("s3://%s/%s", s3s.config.Bucket, *obj.Key)

		backups = append(backups, metadata)
	}

	// Sort by timestamp (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// Delete removes a backup from S3
func (s3s *S3Storage) Delete(ctx context.Context, key string) error {
	s3Key := s3s.buildS3Key(key)

	// Delete object with retry logic
	var err error
	for attempt := 0; attempt < S3MaxRetries; attempt++ {
		_, err = s3s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s3s.config.Bucket),
			Key:    aws.String(s3Key),
		})
		
		if err == nil {
			break
		}
		
		if attempt < S3MaxRetries-1 {
			delay := time.Duration(attempt+1) * S3RetryDelay
			s3s.logger.Warn("S3 DeleteObject attempt %d failed, retrying in %v: %v", 
				attempt+1, delay, err)
			time.Sleep(delay)
		}
	}
	
	if err != nil {
		return fmt.Errorf("failed to delete S3 object after %d attempts: %w", 
			S3MaxRetries, err)
	}

	s3s.logger.Info("Backup deleted successfully from S3: %s", key)
	return nil
}

// Exists checks if a backup exists in S3
func (s3s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	s3Key := s3s.buildS3Key(key)

	// Check object existence with retry logic
	var err error
	for attempt := 0; attempt < S3MaxRetries; attempt++ {
		_, err = s3s.client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(s3s.config.Bucket),
			Key:    aws.String(s3Key),
		})
		
		if err == nil {
			return true, nil
		}
		
		// Check if it's a "not found" error
		var noSuchKey *s3types.NoSuchKey
		var notFound *s3types.NotFound
		if errors.As(err, &noSuchKey) || errors.As(err, &notFound) {
			return false, nil
		}
		
		if attempt < S3MaxRetries-1 {
			delay := time.Duration(attempt+1) * S3RetryDelay
			s3s.logger.Warn("S3 HeadObject attempt %d failed, retrying in %v: %v", 
				attempt+1, delay, err)
			time.Sleep(delay)
		}
	}
	
	return false, fmt.Errorf("failed to check S3 object existence after %d attempts: %w", 
		S3MaxRetries, err)
}

// GetType returns the storage backend type identifier
func (s3s *S3Storage) GetType() string {
	return "s3"
}

// Cleanup performs any necessary cleanup operations
func (s3s *S3Storage) Cleanup(ctx context.Context) error {
	// For S3 storage, cleanup is handled by retention policies
	// This method is here for interface compliance
	return nil
}

// validateS3Access validates S3 connectivity and permissions
func (s3s *S3Storage) validateS3Access(ctx context.Context) error {
	// Check if bucket exists and is accessible
	_, err := s3s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s3s.config.Bucket),
	})
	if err != nil {
		return fmt.Errorf("cannot access S3 bucket %s: %w", s3s.config.Bucket, err)
	}

	// Test write permissions by creating a test object
	testKey := s3s.buildS3Key("test-connectivity")
	testData := []byte("tf-safe connectivity test")
	
	_, err = s3s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s3s.config.Bucket),
		Key:    aws.String(testKey),
		Body:   bytes.NewReader(testData),
	})
	if err != nil {
		return fmt.Errorf("cannot write to S3 bucket %s: %w", s3s.config.Bucket, err)
	}

	// Clean up test object
	_, err = s3s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s3s.config.Bucket),
		Key:    aws.String(testKey),
	})
	if err != nil {
		s3s.logger.Warn("Failed to clean up test object: %v", err)
	}

	return nil
}

// buildS3Key constructs the S3 object key from a backup key
func (s3s *S3Storage) buildS3Key(key string) string {
	if s3s.config.Prefix != "" {
		return fmt.Sprintf("%s%s%s", s3s.config.Prefix, key, BackupFileExtension)
	}
	return fmt.Sprintf("%s%s", key, BackupFileExtension)
}

// extractBackupKey extracts the backup key from an S3 object key
func (s3s *S3Storage) extractBackupKey(s3Key string) string {
	// Remove prefix if present
	key := s3Key
	if s3s.config.Prefix != "" && strings.HasPrefix(key, s3s.config.Prefix) {
		key = strings.TrimPrefix(key, s3s.config.Prefix)
	}
	
	// Remove backup file extension
	if strings.HasSuffix(key, BackupFileExtension) {
		key = strings.TrimSuffix(key, BackupFileExtension)
	}
	
	return key
}

// parseS3Metadata parses backup metadata from S3 object metadata
func (s3s *S3Storage) parseS3Metadata(s3Metadata map[string]string, key string) (*tftypes.BackupMetadata, error) {
	metadata := &tftypes.BackupMetadata{
		ID: key,
	}

	// Parse timestamp
	if timestampStr, ok := s3Metadata[S3MetadataPrefix+"timestamp"]; ok {
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp format: %w", err)
		}
		metadata.Timestamp = timestamp
	} else {
		// Fallback to current time if timestamp is missing
		metadata.Timestamp = time.Now()
	}

	// Parse checksum
	if checksum, ok := s3Metadata[S3MetadataPrefix+"checksum"]; ok {
		metadata.Checksum = checksum
	}

	// Parse encrypted flag
	if encryptedStr, ok := s3Metadata[S3MetadataPrefix+"encrypted"]; ok {
		metadata.Encrypted = encryptedStr == "true"
	}

	return metadata, nil
}

// regularUpload performs a regular S3 upload for smaller files
func (s3s *S3Storage) regularUpload(ctx context.Context, s3Key string, data []byte, s3Metadata map[string]string) error {
	var err error
	for attempt := 0; attempt < S3MaxRetries; attempt++ {
		_, err = s3s.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:   aws.String(s3s.config.Bucket),
			Key:      aws.String(s3Key),
			Body:     bytes.NewReader(data),
			Metadata: s3Metadata,
		})
		
		if err == nil {
			s3s.logger.Info("Backup stored successfully in S3: %s (size: %d bytes)", 
				s3Key, len(data))
			return nil
		}
		
		if attempt < S3MaxRetries-1 {
			delay := time.Duration(attempt+1) * S3RetryDelay
			s3s.logger.Warn("S3 PutObject attempt %d failed, retrying in %v: %v", 
				attempt+1, delay, err)
			time.Sleep(delay)
		}
	}
	
	return fmt.Errorf("failed to upload to S3 after %d attempts: %w", S3MaxRetries, err)
}

// multipartUpload performs a multipart S3 upload for larger files
func (s3s *S3Storage) multipartUpload(ctx context.Context, s3Key string, data []byte, s3Metadata map[string]string) error {
	// Create multipart upload
	createOutput, err := s3s.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:   aws.String(s3s.config.Bucket),
		Key:      aws.String(s3Key),
		Metadata: s3Metadata,
	})
	if err != nil {
		return fmt.Errorf("failed to create multipart upload: %w", err)
	}

	uploadID := createOutput.UploadId
	var completedParts []s3types.CompletedPart
	
	// Upload parts
	partSize := S3MultipartThreshold
	partNumber := int32(1)
	
	for offset := 0; offset < len(data); offset += partSize {
		end := offset + partSize
		if end > len(data) {
			end = len(data)
		}
		
		partData := data[offset:end]
		
		uploadOutput, err := s3s.client.UploadPart(ctx, &s3.UploadPartInput{
			Bucket:     aws.String(s3s.config.Bucket),
			Key:        aws.String(s3Key),
			PartNumber: &partNumber,
			UploadId:   uploadID,
			Body:       bytes.NewReader(partData),
		})
		if err != nil {
			// Abort multipart upload on error
			s3s.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(s3s.config.Bucket),
				Key:      aws.String(s3Key),
				UploadId: uploadID,
			})
			return fmt.Errorf("failed to upload part %d: %w", partNumber, err)
		}
		
		completedParts = append(completedParts, s3types.CompletedPart{
			ETag:       uploadOutput.ETag,
			PartNumber: &partNumber,
		})
		
		partNumber++
	}
	
	// Complete multipart upload
	_, err = s3s.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s3s.config.Bucket),
		Key:      aws.String(s3Key),
		UploadId: uploadID,
		MultipartUpload: &s3types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		// Abort multipart upload on error
		s3s.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
			Bucket:   aws.String(s3s.config.Bucket),
			Key:      aws.String(s3Key),
			UploadId: uploadID,
		})
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}
	
	s3s.logger.Info("Backup stored successfully in S3 using multipart upload: %s (size: %d bytes)", 
		s3Key, len(data))
	return nil
}