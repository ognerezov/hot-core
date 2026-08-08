package tools

import (
	"testing"
)

func TestDetectMediaType(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantMime string
		wantType FileType
	}{
		{
			name:     "JPEG",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
			wantMime: "image/jpeg",
			wantType: Image,
		},
		{
			name:     "PNG",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			wantMime: "image/png",
			wantType: Image,
		},
		{
			name:     "HEIC (Apple Image)",
			data:     []byte{0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70, 0x68, 0x65, 0x69, 0x63},
			wantMime: "image/heic",
			wantType: Image,
		},
		{
			name:     "MP4 Video",
			data:     []byte{0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70, 0x69, 0x73, 0x6F, 0x6D},
			wantMime: "video/mp4",
			wantType: Video,
		},
		{
			name:     "QuickTime MOV",
			data:     []byte{0x00, 0x00, 0x00, 0x14, 0x66, 0x74, 0x79, 0x70, 0x71, 0x74, 0x20, 0x20},
			wantMime: "video/quicktime",
			wantType: Video,
		},
		{
			name:     "AMR Audio",
			data:     []byte("#!AMR\n"),
			wantMime: "audio/amr",
			wantType: Audio,
		},
		{
			name:     "Plain Text",
			data:     []byte("This is plain text"),
			wantMime: "text/plain",
			wantType: Other,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mime, ftype := DetectMediaType(tt.data)
			if mime != tt.wantMime {
				t.Errorf("DetectMediaType() mime = %v, want %v", mime, tt.wantMime)
			}
			if ftype != tt.wantType {
				t.Errorf("DetectMediaType() type = %v, want %v", ftype, tt.wantType)
			}
		})
	}
}
