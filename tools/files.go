package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ognerezov/hot-core/console"
)

var (
	// AllowedHiddenFiles lists hidden files that are allowed in an "empty" directory.
	AllowedHiddenFiles = []string{".DS_Store"}
)

// CloseFile closes an *os.File and logs an error if it fails.
func CloseFile(f *os.File) {
	err := f.Close()
	if err != nil {
		console.RedPrintln(err.Error())
	}
}

// CloseAny closes an io.Closer and logs an error if it fails.
func CloseAny(f io.Closer) {
	err := f.Close()
	if err != nil {
		console.RedPrintln(err.Error())
	}
}

// ReadFile reads the entire content of a file.
func ReadFile(filePath string) ([]byte, error) {
	f, err := os.OpenFile(filepath.FromSlash(filePath), os.O_RDONLY, 0)
	if err != nil {
		console.RedPrintln(err.Error())
		return nil, err
	}
	defer CloseFile(f)
	bytes, err := io.ReadAll(f)
	if err != nil {
		console.RedPrintln(err.Error())
		return nil, err
	}
	return bytes, nil
}

// DoesFileExist checks if a file exists and is readable.
func DoesFileExist(filePath string) bool {
	f, err := os.OpenFile(filepath.FromSlash(filePath), os.O_RDONLY, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	defer CloseFile(f)
	return true
}

// AnyFromFile reads a JSON file and unmarshals it into the provided output structure.
func AnyFromFile[T any](filename string, out *T) error {
	f, err := os.OpenFile(filepath.FromSlash(filename), os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err = f.Close()
		if err != nil {
			console.RedPrintln(err.Error())
		}
	}(f)

	decoder := json.NewDecoder(f)
	return decoder.Decode(out)
}

// AnyFromString unmarshals a JSON string into the provided output structure.
func AnyFromString[T any](raw *string, out *T) error {
	return json.Unmarshal([]byte(*raw), &out)
}

// Save writes bytes to a file, creating it if necessary, with 0755 permissions.
func Save(file string, bytes []byte) error {
	f, err := os.OpenFile(filepath.FromSlash(file), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer CloseFile(f)
	_, err = f.Write(bytes)
	return err
}

// SaveJavascript saves data as a Javascript module (export default).
func SaveJavascript(path string, data any) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	str := "export default\n" + string(bytes)
	err = Save(path, []byte(str))
	if err != nil {
		return err
	}
	return nil
}

// IsEmptyDir checks if a directory is empty (ignoring AllowedHiddenFiles).
func IsEmptyDir(name string) error {
	f, err := os.Open(filepath.FromSlash(name))
	if err != nil {
		return err
	}
	defer CloseFile(f)

	files, err := f.ReadDir(1)
	if len(files) != 0 {
		for _, fileName := range AllowedHiddenFiles {
			if fileName == files[0].Name() {
				return nil
			}
		}
		return fmt.Errorf("%s is not empty", name)
	}
	return nil
}

// Exists checks if a file or directory exists.
func Exists(path string) (bool, error) {
	_, err := os.Stat(filepath.FromSlash(path))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// MkDir creates a directory and all its parent directories.
func MkDir(path string) error {
	return os.MkdirAll(filepath.FromSlash(path), os.ModePerm)
}

// ValidateFilename checks if a filename is safe and valid.
func ValidateFilename(name string) error {
	if name == "" {
		return errors.New("filename is required")
	}
	if len(name) > 200 {
		return errors.New("filename is too long")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return errors.New("filename must not contain path separators")
	}
	if strings.Contains(name, "..") {
		return errors.New("filename must not contain '..'")
	}
	if strings.HasPrefix(name, ".") {
		return errors.New("filename must not start with '.'")
	}
	return nil
}

// GetPathSegment returns the segment of a path at the given index.
// Supports negative indices for counting from the end.
func GetPathSegment(path *string, index int) *string {
	segments := strings.Split(*path, "/")
	_index := index
	if index < 0 {
		_index = len(segments) + index
	}
	if _index < 0 || _index >= len(segments) {
		return nil
	}
	return &segments[_index]
}

// ParseS3Url extracts the bucket and key from a standard S3 URL.
func ParseS3Url(s3Url string) (bucket *string, key *string) {
	u := strings.TrimPrefix(s3Url, "s3://")

	parts := strings.SplitN(u, "/", 2)
	if len(parts) == 2 {
		bucket = &parts[0]
		key = &parts[1]
		return bucket, key
	}

	return &parts[0], nil
}

// S3Url formats a bucket and key into an S3 URI (s3://bucket/key).
func S3Url(bucket, key string) *string {
	// Ensure key doesn't start with / to avoid double slashes
	cleanKey := strings.TrimPrefix(key, "/")
	res := fmt.Sprintf("s3://%s/%s", bucket, cleanKey)
	return &res
}

// GenerateOriginalS3Key generates a unique S3 key for an original upload.
func GenerateOriginalS3Key(userID string, fileName string) string {
	now := time.Now()

	cleanName := filepath.Base(fileName)
	cleanName = strings.ReplaceAll(cleanName, " ", "_")

	return fmt.Sprintf("%s/%s/%s",
		userID,
		now.Format("2006/01/02/15-04"),
		cleanName,
	)
}

// ParseUploadKey extracts metadata from a generated S3 key.
func ParseUploadKey(key string) (userID *string, fileName *string, uploadTime *string) {
	parts := strings.Split(key, "/")
	if len(parts) < 6 {
		return nil, nil, nil
	}
	fileName = &parts[len(parts)-1]
	timeStr := strings.Join(parts[1:5], "/") + parts[5]
	return &parts[0], &timeStr, fileName
}
