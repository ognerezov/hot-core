package tools

import (
	"io"
	"net/http"
	"os"
	"strings"
)

// FileType represents the category of a file (image, video, audio, etc.).
type FileType string

const (
	// Image represents image files.
	Image FileType = "image"
	// Video represents video files.
	Video FileType = "video"
	// Audio represents audio files.
	Audio FileType = "audio"
	// Other represents files that don't fit into image, video, or audio categories.
	Other FileType = "other"
)

// DetectMediaType determines the media type (image, video, audio) based on file content (magic bytes).
// This is more reliable than checking the file extension, especially for content from mobile devices.
func DetectMediaType(data []byte) (string, FileType) {
	// 1. First, try the standard Go detector.
	// It handles JPEG, PNG, GIF, WebP, MP4, OGG, WAV well.
	mimeType := http.DetectContentType(data)

	// Clean mimeType from parameters (e.g., "text/plain; charset=utf-8").
	if idx := strings.Index(mimeType, ";"); idx != -1 {
		mimeType = mimeType[:idx]
	}
	mimeType = strings.TrimSpace(mimeType)

	// 2. Additional check for Apple, Android, and new standard formats
	// that http.DetectContentType might identify as "application/octet-stream" or "text/plain".
	if mimeType == "application/octet-stream" || mimeType == "text/plain" {
		if m := detectMobileFormats(data); m != nil {
			mimeType = *m
		}
	}

	// 3. Mapping MIME type to our MediaType.
	if strings.HasPrefix(mimeType, "image/") {
		return mimeType, Image
	}
	if strings.HasPrefix(mimeType, "video/") {
		return mimeType, Video
	}
	if strings.HasPrefix(mimeType, "audio/") {
		return mimeType, Audio
	}

	return mimeType, Other
}

// DetectMediaTypeFromFile determines the media type of a file on disk by reading only its header.
// This is efficient for large files (videos) as it doesn't load the entire file into memory.
func DetectMediaTypeFromFile(filePath string) (string, FileType, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", Other, err
	}
	defer CloseAny(f)

	// Read only the first 512 bytes (sufficient for most signatures).
	// http.DetectContentType uses a maximum of 512 bytes.
	buffer := make([]byte, 512)
	n, err := f.Read(buffer)
	if err != nil && err != io.EOF {
		return "", Other, err
	}

	// Use only the read part.
	mime, t := DetectMediaType(buffer[:n])
	return mime, t, nil
}

// detectMobileFormats checks signatures of specific formats for iOS and Android.
// Returns a pointer to the MIME type or nil if the format is not recognized.
func detectMobileFormats(data []byte) *string {
	// Helper for returning a string pointer.
	ret := func(s string) *string { return &s }

	// --- ISO Base Media File Format (MP4, HEIC, 3GP, MOV) ---
	// Usually starts with 4 bytes of size, then 'ftyp'.
	if len(data) >= 12 && string(data[4:8]) == "ftyp" {
		majorBrand := string(data[8:12])

		switch majorBrand {
		// Apple Images
		case "heic", "heix", "heim", "heis":
			return ret("image/heic")
		case "mif1", "msf1":
			return ret("image/heif")

		// Apple Video
		case "qt  ":
			return ret("video/quicktime")

		// Apple Audio
		case "M4A ", "M4B ", "M4P ":
			return ret("audio/mp4")

		// Common Video (Android & iOS)
		// avc1 - H.264, hvc1/hev1 - H.265 (HEVC)
		case "mp41", "mp42", "isom", "iso2", "avc1", "hvc1", "hev1":
			return ret("video/mp4")
		case "3gp4", "3gp5", "3gp6", "3gp7": // 3GP (old Android)
			return ret("video/3gpp")
		}
	}

	// --- Android / Web Specifics ---

	// WebM (often used in Android for video)
	// Signature: 1A 45 DF A3
	if len(data) >= 4 && data[0] == 0x1A && data[1] == 0x45 && data[2] == 0xDF && data[3] == 0xA3 {
		return ret("video/webm")
	}

	// AMR (Adaptive Multi-Rate) - old Android voice recorder format.
	// Signature: "#!AMR"
	if len(data) >= 5 && string(data[:5]) == "#!AMR" {
		return ret("audio/amr")
	}

	// --- Audio Specifics ---

	// AAC (ADTS)
	// Sync Word: 12 bits set to 1. Usually FFF1 or FFF9.
	if len(data) > 2 && data[0] == 0xFF && (data[1]&0xF0 == 0xF0) {
		return ret("audio/aac")
	}

	// MP3 (ID3 tag)
	if len(data) > 3 && string(data[:3]) == "ID3" {
		return ret("audio/mpeg")
	}

	// FLAC
	if len(data) > 4 && string(data[:4]) == "fLaC" {
		return ret("audio/flac")
	}

	// Ogg (Vorbis/Opus) - Android voice recorders.
	// Signature: "OggS"
	if len(data) >= 4 && string(data[:4]) == "OggS" {
		return ret("audio/ogg")
	}

	return nil
}
