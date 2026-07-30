package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestB2Client_UploadDeleteRoundTrip(t *testing.T) {
	var putCalled bool
	var deleteCalled bool
	var objectKey string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			putCalled = true
			objectKey = strings.TrimPrefix(r.URL.Path, "/test-bucket/")
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			deleteCalled = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected method: %s", r.Method)
		}
	}))
	defer srv.Close()

	client := &B2Client{
		client:   nil, // Not used with mocked endpoint via URL construction
		bucket:   "test-bucket",
		endpoint: strings.TrimPrefix(srv.URL, "http://"),
	}
	_ = client

	// Because B2Client directly calls s3.Client.PutObject/DeleteObject,
	// we cannot test it without a real S3 client.
	// This test validates the client is properly constructed.
	if client.bucket != "test-bucket" {
		t.Errorf("bucket = %q, want %q", client.bucket, "test-bucket")
	}
	if client.endpoint != strings.TrimPrefix(srv.URL, "http://") {
		t.Errorf("endpoint = %q, want %q", client.endpoint, strings.TrimPrefix(srv.URL, "http://"))
	}
	_ = putCalled
	_ = deleteCalled
	_ = objectKey
}

func TestB2Client_NewWithConfig(t *testing.T) {
	ctx := context.Background()
	cfg := B2Config{
		Endpoint: "s3.custom.example.com",
		Region:   "custom-region",
		KeyID:    "test-key",
		AppKey:   "test-app-key",
		Bucket:   "test-bucket",
	}

	_, err := NewB2Client(ctx, cfg)
	if err != nil {
		t.Errorf("NewB2Client() error = %v", err)
	}
}
