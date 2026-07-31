//go:build integration

package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/minio"
)

const (
	testBucket = "test-bucket"
	minioUser  = "minioadmin"
	minioPass  = "minioadmin"
)

var (
	testClient Client
	verifyS3   *s3.Client
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := minio.Run(ctx, "minio/minio:latest",
		minio.WithUsername(minioUser),
		minio.WithPassword(minioPass),
		testcontainers.WithTmpfs(map[string]string{"/data": "rw"}),
	)
	if err != nil {
		fmt.Printf("start minio container: %v\n", err)
		os.Exit(1)
	}

	endpoint, err := container.ConnectionString(ctx)
	if err != nil {
		fmt.Printf("minio connection string: %v\n", err)
		os.Exit(1)
	}
	if !strings.Contains(endpoint, "://") {
		endpoint = "http://" + endpoint
	}

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: endpoint, HostnameImmutable: true}, nil
		})),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(minioUser, minioPass, "")),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		fmt.Printf("aws config: %v\n", err)
		os.Exit(1)
	}
	verifyS3 = s3.NewFromConfig(awsCfg)

	if _, err := verifyS3.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(testBucket)}); err != nil {
		fmt.Printf("create bucket: %v\n", err)
		os.Exit(1)
	}

	testClient, err = NewB2Client(ctx, B2Config{
		Endpoint: endpoint,
		Region:   "us-east-1",
		KeyID:    minioUser,
		AppKey:   minioPass,
		Bucket:   testBucket,
	})
	if err != nil {
		fmt.Printf("new b2 client: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	_ = container.Terminate(ctx)
	os.Exit(code)
}

func TestB2Client_Upload_CreatesObject(t *testing.T) {
	ctx := context.Background()
	data := []byte("fake-image-bytes")

	path, err := testClient.Upload(ctx, "berita/b1/photo.png", "image/png", bytes.NewReader(data))
	require.NoError(t, err)
	require.Equal(t, "berita/b1/photo.png", path)

	obj, err := verifyS3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(testBucket),
		Key:    aws.String(path),
	})
	require.NoError(t, err)
	defer obj.Body.Close()

	got, err := io.ReadAll(obj.Body)
	require.NoError(t, err)
	require.Equal(t, data, got)
}

func TestB2Client_Upload_TrimsLeadingSlash(t *testing.T) {
	ctx := context.Background()

	path, err := testClient.Upload(ctx, "/berita/b2/photo.png", "image/png", bytes.NewReader([]byte("x")))
	require.NoError(t, err)
	require.Equal(t, "berita/b2/photo.png", path)

	obj, err := verifyS3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(testBucket),
		Key:    aws.String(path),
	})
	require.NoError(t, err)
	obj.Body.Close()
}

func TestB2Client_PresignGet_ServesBytes(t *testing.T) {
	ctx := context.Background()
	data := []byte("presigned-image-bytes")

	path, err := testClient.Upload(ctx, "berita/b1/presigned.png", "image/png", bytes.NewReader(data))
	require.NoError(t, err)

	signedURL, err := testClient.PresignGet(ctx, path, 5*time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, signedURL)

	resp, err := http.Get(signedURL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	got, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, data, got)
}

func TestB2Client_Delete_RemovesObject(t *testing.T) {
	ctx := context.Background()

	path, err := testClient.Upload(ctx, "berita/b1/to-delete.png", "image/png", bytes.NewReader([]byte("x")))
	require.NoError(t, err)

	require.NoError(t, testClient.Delete(ctx, path))

	_, err = verifyS3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(testBucket),
		Key:    aws.String(path),
	})
	var noSuchKey *types.NoSuchKey
	require.ErrorAs(t, err, &noSuchKey)
}
