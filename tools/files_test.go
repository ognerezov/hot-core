package tools

import (
	"testing"
)

func TestGetPathSegment(t *testing.T) {
	path := "users/123/photos/my_pic.jpg"
	tests := []struct {
		name  string
		index int
		want  string
	}{
		{"First", 0, "users"},
		{"Second", 1, "123"},
		{"Last", -1, "my_pic.jpg"},
		{"Second to last", -2, "photos"},
		{"Out of range", 5, ""},
		{"Out of range negative", -5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPathSegment(&path, tt.index)
			if tt.want == "" {
				if got != nil {
					t.Errorf("GetPathSegment() = %v, want nil", *got)
				}
			} else {
				if got == nil || *got != tt.want {
					if got == nil {
						t.Errorf("GetPathSegment() = nil, want %v", tt.want)
					} else {
						t.Errorf("GetPathSegment() = %v, want %v", *got, tt.want)
					}
				}
			}
		})
	}
}

func TestS3Url(t *testing.T) {
	bucket := "my-bucket"
	key := "/my/key.jpg"
	want := "s3://my-bucket/my/key.jpg"
	got := S3Url(bucket, key)
	if *got != want {
		t.Errorf("S3Url() = %v, want %v", *got, want)
	}
}

func TestValidateFilename(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid", "photo.jpg", false},
		{"Empty", "", true},
		{"Too long", string(make([]byte, 201)), true},
		{"Path separator slash", "path/to/file", true},
		{"Path separator backslash", "path\\to\\file", true},
		{"Dot dot", "dir/../file", true},
		{"Starts with dot", ".hidden", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilename(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilename() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
