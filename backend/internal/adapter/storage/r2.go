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
	PrefixDisputes = "disputes"
)

// R2Storage wraps an S3-compatible client configured for Cloudflare R2.
type R2Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	publicURL     string // base URL for public file access (e.g. https://cdn.seva.io)
}

// NewR2Storage creates a new Cloudflare R2 storage client.
// The S3-compatible endpoint is derived from the accountID.
// publicURL is the base URL used to construct public file URLs (e.g. a custom
// domain or an R2.dev subdomain).
func NewR2Storage(accountID, accessKeyID, accessKeySecret, bucket, publicURL string) *R2Storage {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	cfg := aws.Config{
		Region: "auto", // R2 ignores region but the SDK requires one
		Credentials: credentials.NewStaticCredentialsProvider(
			accessKeyID,
			accessKeySecret,
			"",
		),
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	presignClient := s3.NewPresignClient(client)

	log.Info().
		Str("bucket", bucket).
		Str("endpoint", endpoint).
		Msg("R2 storage client initialized")

	return &R2Storage{
		client:        client,
		presignClient: presignClient,
		bucket:        bucket,
		publicURL:     publicURL,
	}
}

// Upload stores a file in the R2 bucket at the given key and returns the
// public URL of the uploaded object.
//
// Key organization:
//   - kyc/{userID}/{filename}       -- KYC documents
//   - jobs/{jobID}/{filename}       -- job-related media
//   - profiles/{userID}/avatar      -- profile avatars
//   - disputes/{disputeID}/{filename} -- dispute evidence
func (s *R2Storage) Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        data,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPrivate,
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("r2 upload %s: %w", key, err)
	}

	url := s.objectURL(key)

	log.Debug().
		Str("key", key).
		Str("content_type", contentType).
		Str("bucket", s.bucket).
		Msg("file uploaded to R2")

	return url, nil
}

// Download retrieves a file from the R2 bucket. The caller is responsible for
// closing the returned ReadCloser.
func (s *R2Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	output, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("r2 download %s: %w", key, err)
	}

	log.Debug().
		Str("key", key).
		Str("bucket", s.bucket).
		Msg("file downloaded from R2")

	return output.Body, nil
}

// Delete removes a file from the R2 bucket.
func (s *R2Storage) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("r2 delete %s: %w", key, err)
	}

	log.Debug().
		Str("key", key).
		Str("bucket", s.bucket).
		Msg("file deleted from R2")

	return nil
}

// GeneratePresignedURL creates a pre-signed URL that allows temporary direct
// upload or download access to a private object.
func (s *R2Storage) GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	presignResult, err := s.presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("r2 presign %s: %w", key, err)
	}

	log.Debug().
		Str("key", key).
		Dur("expiry", expiry).
		Msg("presigned URL generated")

	return presignResult.URL, nil
}

// objectURL returns the public URL for an object using the configured publicURL
// base. For example: https://cdn.seva.io/kyc/user123/doc.pdf
func (s *R2Storage) objectURL(key string) string {
	return fmt.Sprintf("%s/%s", s.publicURL, key)
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

// DisputeEvidenceKey constructs the storage key for dispute evidence.
func DisputeEvidenceKey(disputeID, filename string) string {
	return fmt.Sprintf("%s/%s/%s", PrefixDisputes, disputeID, filename)
}
