package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"
)

// StorageProvider defines the interface for object storage operations.
type StorageProvider interface {
	Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

// Key prefix constants for organized storage.
const (
	PrefixKYC      = "kyc"
	PrefixJobs     = "jobs"
	PrefixProfiles = "profiles"
)

// StorageClient wraps an S3-compatible client for Cloudflare R2 or AWS S3.
type StorageClient struct {
	client         *s3.Client
	presignClient  *s3.PresignClient
	bucket         string
	endpoint       string
}

// NewStorageClient creates a new S3-compatible storage client configured for
// Cloudflare R2 or any S3-compatible endpoint.
func NewStorageClient(bucket, region, accessKey, secretKey, endpoint string) *StorageClient {
	cfg := aws.Config{
		Region: region,
		Credentials: credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		),
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true
	})

	presignClient := s3.NewPresignClient(client)

	return &StorageClient{
		client:        client,
		presignClient: presignClient,
		bucket:        bucket,
		endpoint:      endpoint,
	}
}

// Upload stores a file in the bucket at the given key and returns the public URL.
//
// Key organization:
//   - kyc/{userID}/{filename}       — KYC documents
//   - jobs/{jobID}/{filename}       — job-related media
//   - profiles/{userID}/avatar      — profile avatars
func (s *StorageClient) Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        data,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPrivate,
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("storage upload %s: %w", key, err)
	}

	url := s.objectURL(key)

	log.Debug().
		Str("key", key).
		Str("content_type", contentType).
		Str("bucket", s.bucket).
		Msg("file uploaded to storage")

	return url, nil
}

// Download retrieves a file from the bucket. The caller is responsible for
// closing the returned ReadCloser.
func (s *StorageClient) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	output, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("storage download %s: %w", key, err)
	}

	log.Debug().
		Str("key", key).
		Str("bucket", s.bucket).
		Msg("file downloaded from storage")

	return output.Body, nil
}

// Delete removes a file from the bucket.
func (s *StorageClient) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("storage delete %s: %w", key, err)
	}

	log.Debug().
		Str("key", key).
		Str("bucket", s.bucket).
		Msg("file deleted from storage")

	return nil
}

// GeneratePresignedURL creates a pre-signed URL that allows temporary access
// to a private object.
func (s *StorageClient) GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	presignResult, err := s.presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("storage presign %s: %w", key, err)
	}

	log.Debug().
		Str("key", key).
		Dur("expiry", expiry).
		Msg("presigned URL generated")

	return presignResult.URL, nil
}

// objectURL returns the public URL for an object. For R2, this uses the
// configured endpoint; for standard S3, it constructs the virtual-hosted URL.
func (s *StorageClient) objectURL(key string) string {
	if s.endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key)
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, key)
}

// KYCKey constructs the storage key for a KYC document.
func KYCKey(userID, filename string) string {
	return fmt.Sprintf("%s/%s/%s", PrefixKYC, userID, filename)
}

// JobKey constructs the storage key for a job-related file.
func JobKey(jobID, filename string) string {
	return fmt.Sprintf("%s/%s/%s", PrefixJobs, jobID, filename)
}

// ProfileAvatarKey constructs the storage key for a user's profile avatar.
func ProfileAvatarKey(userID string) string {
	return fmt.Sprintf("%s/%s/avatar", PrefixProfiles, userID)
}
