package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type B2Config struct {
	Endpoint string
	Region   string
	KeyID    string
	AppKey   string
	Bucket   string
}

type B2Client struct {
	client   *s3.Client
	bucket   string
	endpoint string
}

func NewB2Client(ctx context.Context, cfg B2Config) (*B2Client, error) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               fmt.Sprintf("https://%s", cfg.Endpoint),
			HostnameImmutable: true,
		}, nil
	})

	region := cfg.Region
	if region == "" {
		region = "us-east-005"
	}

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.KeyID, cfg.AppKey, "")),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}
	awsCfg.HTTPClient = httpClient

	return &B2Client{
		client:   s3.NewFromConfig(awsCfg),
		bucket:   cfg.Bucket,
		endpoint: cfg.Endpoint,
	}, nil
}

func (c *B2Client) Upload(ctx context.Context, objectPath string, contentType string, reader io.Reader) (string, error) {
	clean := strings.TrimPrefix(objectPath, "/")
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(clean),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("s3 upload: %w", err)
	}

	publicURL := fmt.Sprintf("https://%s.%s/%s", c.bucket, c.endpoint, clean)
	return publicURL, nil
}

func (c *B2Client) Delete(ctx context.Context, objectPath string) error {
	clean := strings.TrimPrefix(objectPath, "/")
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(clean),
	})
	if err != nil {
		return fmt.Errorf("s3 delete: %w", err)
	}
	return nil
}
