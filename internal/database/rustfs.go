package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/sky-flux/cms/internal/config"
)

// NewRustFS creates an S3-compatible client for RustFS object storage.
func NewRustFS(cfg *config.Config) (*s3.Client, error) {
	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(cfg.RustFS.Endpoint),
		Region:       cfg.RustFS.Region,
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(
				cfg.RustFS.AccessKey,
				cfg.RustFS.SecretKey,
				"",
			),
		),
		UsePathStyle: true,
	})

	// Verify connectivity by listing buckets
	_, err := client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("rustfs connection failed: %w", err)
	}

	return client, nil
}
