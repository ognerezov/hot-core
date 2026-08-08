package aws

import (
	"testing"
)

func TestParseS3Uri(t *testing.T) {
	tests := []struct {
		name       string
		uri        string
		wantBucket string
		wantKey    string
		wantErr    bool
	}{
		{
			name:       "Valid URI",
			uri:        "s3://my-bucket/my/key/file.jpg",
			wantBucket: "my-bucket",
			wantKey:    "my/key/file.jpg",
			wantErr:    false,
		},
		{
			name:       "Valid URI with root file",
			uri:        "s3://my-bucket/file.jpg",
			wantBucket: "my-bucket",
			wantKey:    "file.jpg",
			wantErr:    false,
		},
		{
			name:       "Invalid Prefix",
			uri:        "http://my-bucket/file.jpg",
			wantBucket: "",
			wantKey:    "",
			wantErr:    true,
		},
		{
			name:       "Missing Key",
			uri:        "s3://my-bucket",
			wantBucket: "",
			wantKey:    "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, key, err := ParseS3Uri(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseS3Uri() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if bucket != tt.wantBucket {
				t.Errorf("ParseS3Uri() bucket = %v, want %v", bucket, tt.wantBucket)
			}
			if key != tt.wantKey {
				t.Errorf("ParseS3Uri() key = %v, want %v", key, tt.wantKey)
			}
		})
	}
}
