package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ognerezov/hot-core/tools"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"
)

const (
	// PreSignedLinksTtl defines the duration in minutes for which presigned links remain valid.
	PreSignedLinksTtl = 30
)

var (
	s3Client *s3.Client
	signer   *s3.PresignClient
	tmClient *transfermanager.Client
)

// SetS3Client sets a custom S3 client and initializes a new presigner.
func SetS3Client(client *s3.Client) {
	s3Client = client
	signer = s3.NewPresignClient(s3Client)
	tmClient = transfermanager.New(s3Client, func(o *transfermanager.Options) {
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	})
}

func getS3Client() *s3.Client {
	if s3Client != nil {
		return s3Client
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(os.Getenv("CDK_REGION")))
	if err != nil {
		panic(fmt.Errorf("failed to load AWS config: %v", err))
	}

	s3Client = s3.NewFromConfig(cfg)
	signer = s3.NewPresignClient(s3Client)
	tmClient = transfermanager.New(s3Client, func(o *transfermanager.Options) {
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	})
	return s3Client
}

func getTMClient() *transfermanager.Client {
	if tmClient != nil {
		return tmClient
	}
	tmClient = transfermanager.New(getS3Client(), func(o *transfermanager.Options) {
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	})
	return tmClient
}

// GetSigner returns a singleton S3 presigner, initializing it if necessary.
func GetSigner() *s3.PresignClient {
	if signer != nil {
		return signer
	}
	getS3Client()
	return signer
}

// PresignUploadRequest represents the request to generate a presigned upload URL.
type PresignUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType,omitempty"`
}

// PresignUploadResponse contains the details for performing a presigned upload.
type PresignUploadResponse struct {
	Bucket        string            `json:"bucket"`
	Key           string            `json:"key"`
	Method        string            `json:"method"`
	URL           string            `json:"url"`
	ExpiresInSec  int64             `json:"expiresInSec"`
	RequiredHeads map[string]string `json:"requiredHeaders,omitempty"`
}

// GetUploadUrl generates a presigned URL for uploading a file to S3.
func GetUploadUrl(bucket string, key string, contentType string) (*PresignUploadResponse, error) {

	preSigner := GetSigner()

	put := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	requiredHeaders := map[string]string{}
	if strings.TrimSpace(contentType) != "" {
		ct := strings.TrimSpace(contentType)
		put.ContentType = &ct
		requiredHeaders["Content-Type"] = ct
	}

	expires := PreSignedLinksTtl * time.Minute
	presigned, err := preSigner.PresignPutObject(context.Background(), put, func(po *s3.PresignOptions) {
		po.Expires = expires
	})
	if err != nil {
		return nil, err
	}

	return &PresignUploadResponse{
		Bucket:        bucket,
		Key:           key,
		Method:        "PUT",
		URL:           presigned.URL,
		ExpiresInSec:  int64(expires.Seconds()),
		RequiredHeads: requiredHeaders,
	}, nil
}

// GetDownloadUrl generates a presigned URL for downloading a file
func GetDownloadUrl(bucket string, key string) (string, error) {
	preSigner := GetSigner()
	expires := PreSignedLinksTtl * time.Minute

	req, err := preSigner.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, func(po *s3.PresignOptions) {
		po.Expires = expires
	})
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// GetDownloadUrlFromUri generates a presigned URL from an S3 URI (s3://bucket/key)
func GetDownloadUrlFromUri(uri string) (string, int64, error) {
	bucket, key, err := ParseS3Uri(uri)
	if err != nil {
		return "", 0, err
	}
	url, err := GetDownloadUrl(bucket, key)
	if err != nil {
		return "", 0, err
	}
	return url, int64(PreSignedLinksTtl * 60), nil
}

// DownloadFromS3 downloads a file from S3 to a local destination path.
func DownloadFromS3(bucket, key, destPath string) error {
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer tools.CloseAny(file)

	tm := getTMClient()
	_, err = tm.DownloadObject(context.Background(), &transfermanager.DownloadObjectInput{
		Bucket:   &bucket,
		Key:      &key,
		WriterAt: file,
	})
	return err
}

// UploadToS3 uploads a local file to S3 with optional metadata.
func UploadToS3(bucket, key, srcPath string, meta *map[string]string) error {
	log.Info().Msgf("Uploading %s to %s/%s", srcPath, bucket, key)
	file, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer tools.CloseAny(file)

	tm := getTMClient()
	contentType, _, _ := tools.DetectMediaTypeFromFile(srcPath)
	input := transfermanager.UploadObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        file,
		ContentType: &contentType,
	}
	if meta != nil {
		input.Metadata = *meta
	}
	_, err = tm.UploadObject(context.Background(), &input)
	return err
}

// UpdateS3Meta updates the metadata of an existing S3 object.
func UpdateS3Meta(bucket, key string, meta *map[string]string) error {
	client := getS3Client()
	_, err := client.CopyObject(context.Background(), &s3.CopyObjectInput{
		Bucket:            &bucket,
		Key:               &key,
		CopySource:        aws.String(bucket + "/" + key),
		MetadataDirective: types.MetadataDirectiveReplace,
		Metadata:          *meta,
	})
	return err
}

// PutS3Tagging adds or updates tags on an existing S3 object.
func PutS3Tagging(bucket, key string, tagMap *map[string]string) error {
	tags := make([]types.Tag, 0)
	for k, v := range *tagMap {
		tags = append(tags, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	client := getS3Client()
	_, err := client.PutObjectTagging(context.Background(), &s3.PutObjectTaggingInput{
		Bucket: &bucket,
		Key:    &key,
		Tagging: &types.Tagging{
			TagSet: tags,
		},
	})
	return err
}

// ParseS3Uri parses an S3 URI in the format s3://bucket/key into bucket and key
func ParseS3Uri(uri string) (bucket, key string, err error) {
	if !strings.HasPrefix(uri, "s3://") {
		return "", "", fmt.Errorf("invalid S3 URI: must start with s3://")
	}

	// Remove s3:// prefix
	path := strings.TrimPrefix(uri, "s3://")

	// Split into bucket and key
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid S3 URI: missing key after bucket")
	}

	bucket = parts[0]
	key = parts[1]

	if bucket == "" {
		return "", "", fmt.Errorf("invalid S3 URI: empty bucket")
	}
	if key == "" {
		return "", "", fmt.Errorf("invalid S3 URI: empty key")
	}

	return bucket, key, nil
}

// GetObjectAsBase64FromUri downloads an object from S3 using an s3:// URI and returns it as a Base64 encoded string.
func GetObjectAsBase64FromUri(ctx context.Context, uri string) (string, error) {
	bucket, key, err := ParseS3Uri(uri)
	if err != nil {
		return "", err
	}
	return GetObjectAsBase64(ctx, bucket, key)
}

// GetObjectAsBase64 downloads an object from S3 and returns it as a Base64 encoded string.
func GetObjectAsBase64(ctx context.Context, bucket, key string) (string, error) {
	tm := getTMClient()

	// Download object
	out, err := tm.GetObject(ctx, &transfermanager.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to download object: %w", err)
	}

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read object body: %w", err)
	}

	// Encode to Base64
	encodedString := base64.StdEncoding.EncodeToString(data)

	return encodedString, nil
}
