package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Store implements Store using an S3-compatible backend.
type S3Store struct {
	client    *s3.Client
	presign   *s3.PresignClient
	bucket    string
	publicURL string
}

// S3Option configures the underlying S3 client.
type S3Option func(*s3.Options)

// NewS3Store creates a new S3-compatible storage backend.
func NewS3Store(ctx context.Context, cfg Config, opts ...S3Option) (*S3Store, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
		for _, opt := range opts {
			opt(o)
		}
	})

	return &S3Store{
		client:    client,
		presign:   s3.NewPresignClient(client),
		bucket:    cfg.Bucket,
		publicURL: strings.TrimRight(cfg.PublicBase, "/"),
	}, nil
}

func (s *S3Store) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("uploading object %q: %w", key, err)
	}

	return nil
}

func (s *S3Store) PresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	req, err := s.presign.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(expiry),
	)
	if err != nil {
		return "", fmt.Errorf("presigning object %q: %w", key, err)
	}

	return req.URL, nil
}

func (s *S3Store) PublicURL(key string) string {
	if s.publicURL == "" {
		panic("PublicURL called on store with no public base URL configured")
	}

	return s.publicURL + "/" + key
}

func (s *S3Store) PresignedPUT(ctx context.Context, key string, contentType string, expiry time.Duration) (string, error) {
	req, err := s.presign.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:      aws.String(s.bucket),
			Key:         aws.String(key),
			ContentType: aws.String(contentType),
		},
		s3.WithPresignExpires(expiry),
	)
	if err != nil {
		return "", fmt.Errorf("presigning PUT for %q: %w", key, err)
	}

	return req.URL, nil
}

func (s *S3Store) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("downloading object %q: %w", key, err)
	}

	return out.Body, nil
}

func (s *S3Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("deleting object %q: %w", key, err)
	}

	return nil
}
