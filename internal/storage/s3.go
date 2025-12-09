package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"seaply/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Storage struct {
	client  *s3.Client
	bucket  string
	baseURL string
	cfg     config.S3Config
}

type UploadResult struct {
	Key      string `json:"key"`
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

// Folder paths for different content types
type FolderType string

const (
	FolderProduct  FolderType = "products"
	FolderSKU      FolderType = "skus"
	FolderBanner   FolderType = "banners"
	FolderPopup    FolderType = "popups"
	FolderProfile  FolderType = "profiles"
	FolderFlag     FolderType = "flags"
	FolderIcon     FolderType = "icons"
	FolderPayment  FolderType = "payment"
	FolderExport   FolderType = "exports"
	FolderCategory FolderType = "categories"
	FolderSection  FolderType = "sections"
)

func NewS3Storage(cfg config.S3Config) (*S3Storage, error) {
	// Create custom endpoint resolver for non-AWS S3 compatible services
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if cfg.Endpoint != "" {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				HostnameImmutable: true,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Load AWS config
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			),
		),
		awsconfig.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true // Required for non-AWS S3 compatible services
	})

	return &S3Storage{
		client:  client,
		bucket:  cfg.Bucket,
		baseURL: cfg.BaseURL,
		cfg:     cfg,
	}, nil
}

// Upload uploads a file to S3
func (s *S3Storage) Upload(ctx context.Context, folder FolderType, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	key := fmt.Sprintf("%s/%s", folder, filename)

	// Detect content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = detectContentType(ext)
	}

	// Upload to S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         "public-read",
	})
	if err != nil {
		return nil, fmt.Errorf("upload to S3: %w", err)
	}

	return &UploadResult{
		Key:      key,
		URL:      s.GetURL(key),
		Filename: filename,
	}, nil
}

// UploadWithOriginalName uploads a file to S3 using the original filename (sanitized)
func (s *S3Storage) UploadWithOriginalName(ctx context.Context, folder FolderType, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	// Sanitize filename: remove path, keep only filename
	originalName := filepath.Base(header.Filename)
	ext := filepath.Ext(originalName)
	nameWithoutExt := strings.TrimSuffix(originalName, ext)

	// Sanitize: keep alphanumeric, dash, underscore, and dots
	var sanitized strings.Builder
	for _, r := range nameWithoutExt {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			sanitized.WriteRune(r)
		} else if r == ' ' {
			sanitized.WriteRune('_')
		}
	}

	sanitizedName := sanitized.String()

	// Limit length to 100 characters
	if len(sanitizedName) > 100 {
		sanitizedName = sanitizedName[:100]
	}

	// Reconstruct filename with extension
	filename := sanitizedName + ext
	if filename == ext || sanitizedName == "" {
		// If name is empty after sanitization, use a default with timestamp
		filename = fmt.Sprintf("image_%d%s", time.Now().Unix(), ext)
	}

	key := fmt.Sprintf("%s/%s", folder, filename)

	// Detect content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = detectContentType(ext)
	}

	// Upload to S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         "public-read",
	})
	if err != nil {
		return nil, fmt.Errorf("upload to S3: %w", err)
	}

	return &UploadResult{
		Key:      key,
		URL:      s.GetURL(key),
		Filename: filename,
	}, nil
}

// UploadFromReader uploads from an io.Reader
func (s *S3Storage) UploadFromReader(ctx context.Context, folder FolderType, reader io.Reader, filename string, contentType string) (*UploadResult, error) {
	ext := filepath.Ext(filename)
	newFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	key := fmt.Sprintf("%s/%s", folder, newFilename)

	if contentType == "" {
		contentType = detectContentType(ext)
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
		ACL:         "public-read",
	})
	if err != nil {
		return nil, fmt.Errorf("upload to S3: %w", err)
	}

	return &UploadResult{
		Key:      key,
		URL:      s.GetURL(key),
		Filename: newFilename,
	}, nil
}

// UploadBytes uploads bytes directly
func (s *S3Storage) UploadBytes(ctx context.Context, folder FolderType, data []byte, filename string, contentType string) (*UploadResult, error) {
	ext := filepath.Ext(filename)
	newFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	key := fmt.Sprintf("%s/%s", folder, newFilename)

	if contentType == "" {
		contentType = detectContentType(ext)
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        strings.NewReader(string(data)),
		ContentType: aws.String(contentType),
		ACL:         "public-read",
	})
	if err != nil {
		return nil, fmt.Errorf("upload to S3: %w", err)
	}

	return &UploadResult{
		Key:      key,
		URL:      s.GetURL(key),
		Filename: newFilename,
	}, nil
}

// Delete removes a file from S3
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete from S3: %w", err)
	}
	return nil
}

// DeleteByURL removes a file by its URL
func (s *S3Storage) DeleteByURL(ctx context.Context, url string) error {
	key := s.GetKeyFromURL(url)
	if key == "" {
		return nil
	}
	return s.Delete(ctx, key)
}

// GetURL returns the full URL for a key
func (s *S3Storage) GetURL(key string) string {
	if s.baseURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.baseURL, "/"), key)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.cfg.Region, key)
}

// GetKeyFromURL extracts the key from a URL
func (s *S3Storage) GetKeyFromURL(url string) string {
	if s.baseURL != "" {
		return strings.TrimPrefix(url, s.baseURL+"/")
	}
	// Handle default S3 URL format
	prefix := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", s.bucket, s.cfg.Region)
	return strings.TrimPrefix(url, prefix)
}

// GetPresignedURL generates a presigned URL for temporary access
func (s *S3Storage) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// Exists checks if a file exists in S3
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if error is "not found"
		return false, nil
	}
	return true, nil
}

// List returns files in a folder
func (s *S3Storage) List(ctx context.Context, folder FolderType, maxKeys int32) ([]string, error) {
	output, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.bucket),
		Prefix:  aws.String(string(folder) + "/"),
		MaxKeys: aws.Int32(maxKeys),
	})
	if err != nil {
		return nil, fmt.Errorf("list S3 objects: %w", err)
	}

	var keys []string
	for _, obj := range output.Contents {
		keys = append(keys, *obj.Key)
	}
	return keys, nil
}

// Copy copies a file within S3
func (s *S3Storage) Copy(ctx context.Context, sourceKey, destFolder FolderType, destFilename string) (*UploadResult, error) {
	destKey := fmt.Sprintf("%s/%s", destFolder, destFilename)
	copySource := fmt.Sprintf("%s/%s", s.bucket, sourceKey)

	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(destKey),
		ACL:        "public-read",
	})
	if err != nil {
		return nil, fmt.Errorf("copy S3 object: %w", err)
	}

	return &UploadResult{
		Key:      destKey,
		URL:      s.GetURL(destKey),
		Filename: destFilename,
	}, nil
}

// Helper to detect content type from extension
func detectContentType(ext string) string {
	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".csv":  "text/csv",
		".json": "application/json",
	}

	if ct, ok := contentTypes[strings.ToLower(ext)]; ok {
		return ct
	}
	return "application/octet-stream"
}

// Validate file type
func ValidateImageFile(header *multipart.FileHeader) error {
	allowedTypes := map[string]bool{
		"image/jpeg":    true,
		"image/png":     true,
		"image/gif":     true,
		"image/webp":    true,
		"image/svg+xml": true,
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		return fmt.Errorf("invalid file type: %s", contentType)
	}

	// Check file size (max 5MB)
	if header.Size > 5*1024*1024 {
		return fmt.Errorf("file too large: max 5MB allowed")
	}

	return nil
}

// Validate export file type
func ValidateExportFile(header *multipart.FileHeader) error {
	allowedTypes := map[string]bool{
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		"text/csv": true,
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		return fmt.Errorf("invalid file type: %s", contentType)
	}

	return nil
}
