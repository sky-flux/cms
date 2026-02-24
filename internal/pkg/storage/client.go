package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Client wraps an S3-compatible client for object storage operations.
type Client struct {
	s3     *s3.Client
	bucket string
	cdnURL string // public URL prefix (e.g., "https://cdn.example.com")
}

// NewClient creates a storage client.
// cdnURL is the public base URL for accessing files (e.g., "http://localhost:9000/cms-media").
func NewClient(s3Client *s3.Client, bucket, cdnURL string) *Client {
	return &Client{s3: s3Client, bucket: bucket, cdnURL: cdnURL}
}

// Available returns true if the S3 client is configured.
func (c *Client) Available() bool {
	return c.s3 != nil
}

// Upload stores an object in S3.
func (c *Client) Upload(ctx context.Context, key string, data io.Reader, contentType string, size int64) error {
	if c.s3 == nil {
		return fmt.Errorf("storage client not available")
	}

	input := &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		Body:          data,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	}

	if _, err := c.s3.PutObject(ctx, input); err != nil {
		return fmt.Errorf("upload %s: %w", key, err)
	}
	return nil
}

// UploadBytes is a convenience wrapper for uploading byte slices.
func (c *Client) UploadBytes(ctx context.Context, key string, data []byte, contentType string) error {
	return c.Upload(ctx, key, bytes.NewReader(data), contentType, int64(len(data)))
}

// Delete removes a single object.
func (c *Client) Delete(ctx context.Context, key string) error {
	if c.s3 == nil {
		return nil
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}

	if _, err := c.s3.DeleteObject(ctx, input); err != nil {
		return fmt.Errorf("delete %s: %w", key, err)
	}
	return nil
}

// BatchDelete removes multiple objects.
func (c *Client) BatchDelete(ctx context.Context, keys []string) error {
	if c.s3 == nil || len(keys) == 0 {
		return nil
	}

	objects := make([]types.ObjectIdentifier, len(keys))
	for i, k := range keys {
		objects[i] = types.ObjectIdentifier{Key: aws.String(k)}
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(c.bucket),
		Delete: &types.Delete{Objects: objects, Quiet: aws.Bool(true)},
	}

	if _, err := c.s3.DeleteObjects(ctx, input); err != nil {
		return fmt.Errorf("batch delete: %w", err)
	}
	return nil
}

// PublicURL returns the public URL for an object key.
func (c *Client) PublicURL(key string) string {
	return c.cdnURL + "/" + key
}

// EnsureBucket creates the bucket if it doesn't exist.
func (c *Client) EnsureBucket(ctx context.Context) error {
	if c.s3 == nil {
		return nil
	}

	_, err := c.s3.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(c.bucket)})
	if err == nil {
		return nil // bucket exists
	}

	_, err = c.s3.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(c.bucket)})
	if err != nil {
		return fmt.Errorf("create bucket %s: %w", c.bucket, err)
	}
	return nil
}
