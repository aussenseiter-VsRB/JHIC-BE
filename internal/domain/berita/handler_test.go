package berita

import (
	"testing"
)

func TestExtractObjectPath(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "simple path",
			url:     "https://jhic-berita-images.s3.eu-central-003.backblazeb2.com/berita/abc/def.jpg",
			want:    "berita/abc/def.jpg",
			wantErr: false,
		},
		{
			name:    "root object",
			url:     "https://bucket.endpoint.com/object.png",
			want:    "object.png",
			wantErr: false,
		},
		{
			name:    "nested path",
			url:     "https://bucket.endpoint.com/a/b/c/d.png",
			want:    "a/b/c/d.png",
			wantErr: false,
		},
		{
			name:    "invalid url",
			url:     "://invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			url:     "",
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractObjectPath(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractObjectPath() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractObjectPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
