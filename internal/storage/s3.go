package storage

import (
	"context"
	"errors"
	"io"
	"time"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Store struct {
	client    *s3.Client
	presign   *s3.PresignClient
	bucket    string
	publicURL string
}

// NewS3Store creates a new S3-compatible storage backend.
func NewS3Store(ctx context.Context, cfg Config) (*S3Store, error) {

	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           cfg.Endpoint,
				SigningRegion: cfg.Region,
			}, nil
		},
	)

	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(cfg.Region),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Store{
		client:    client,
		presign:   s3.NewPresignClient(client),
		bucket:    cfg.Bucket,
		publicURL: cfg.PublicBase,
	}, nil
}

func (s *S3Store) Upload(
	ctx context.Context,
	key string,
	body io.Reader,
	contentType string,
) error {

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})

	return err
}

func (s *S3Store) PresignedURL(
	ctx context.Context,
	key string,
	expiry time.Duration,
) (string, error) {

	req, err := s.presign.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(expiry),
	)

	if err != nil {
		return "", err
	}

	return req.URL, nil
}

func (s *S3Store) PublicURL(key string) (string, error) {

	if s.publicURL == "" {
		return "", errors.New("store has no public base url configured")
	}

	return strings.TrimRight(s.publicURL, "/") + "/" + key, nil
}

func (s *S3Store) Delete(ctx context.Context, key string) error {

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	return err
}