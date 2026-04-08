package services

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/warui/event-ticketing-api/internal/config"
)

type StorageService struct {
	cfg      *config.Config
	s3Client *s3.S3
	useLocal bool
}

func NewStorageService(cfg *config.Config) (*StorageService, error) {
	service := &StorageService{
		cfg:      cfg,
		useLocal: cfg.StorageType == "local" || cfg.AWSAccessKeyID == "" || cfg.AWSBucketName == "",
	}

	if !service.useLocal {
		// Initialize S3/R2 client
		awsConfig := &aws.Config{
			Region:      aws.String(cfg.AWSRegion),
			Credentials: credentials.NewStaticCredentials(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, ""),
		}

		if cfg.AWSEndpoint != "" {
			awsConfig.Endpoint = aws.String(cfg.AWSEndpoint)
			awsConfig.S3ForcePathStyle = aws.Bool(true)
		}

		sess, err := session.NewSession(awsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS session: %w", err)
		}

		service.s3Client = s3.New(sess)
	} else {
		// Ensure local storage directories exist
		dirs := []string{
			filepath.Join(cfg.LocalStoragePath, "events"),
			filepath.Join(cfg.LocalStoragePath, "tickets", "qrcodes"),
			filepath.Join(cfg.LocalStoragePath, "tickets", "pdfs"),
		}

		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}
	}

	return service, nil
}

// UploadFile uploads a file to storage (S3/R2 or local)
func (s *StorageService) UploadFile(data []byte, folder, filename string) (string, error) {
	if s.useLocal {
		return s.uploadLocal(data, folder, filename)
	}
	return s.uploadS3(data, folder, filename)
}

func (s *StorageService) uploadLocal(data []byte, folder, filename string) (string, error) {
	dirPath := filepath.Join(s.cfg.LocalStoragePath, folder)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	filePath := filepath.Join(dirPath, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return relative URL path
	return fmt.Sprintf("/storage/%s/%s", folder, filename), nil
}

func (s *StorageService) uploadS3(data []byte, folder, filename string) (string, error) {
	key := fmt.Sprintf("%s/%s", folder, filename)

	_, err := s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.AWSBucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(getContentType(filename)),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Return public URL
	if s.cfg.AWSEndpoint != "" {
		return fmt.Sprintf("%s/%s/%s", s.cfg.AWSEndpoint, s.cfg.AWSBucketName, key), nil
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.cfg.AWSBucketName, s.cfg.AWSRegion, key), nil
}

// DeleteFile deletes a file from storage
func (s *StorageService) DeleteFile(fileURL string) error {
	if s.useLocal {
		return s.deleteLocal(fileURL)
	}
	return s.deleteS3(fileURL)
}

func (s *StorageService) deleteLocal(fileURL string) error {
	// Extract path from URL
	filePath := filepath.Join(s.cfg.LocalStoragePath, filepath.Base(fileURL))
	return os.Remove(filePath)
}

func (s *StorageService) deleteS3(fileURL string) error {
	// Extract key from URL
	key := filepath.Base(fileURL)

	_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.AWSBucketName),
		Key:    aws.String(key),
	})

	return err
}

// GetFile retrieves a file from storage
func (s *StorageService) GetFile(fileURL string) ([]byte, error) {
	if s.useLocal {
		return s.getLocal(fileURL)
	}
	return s.getS3(fileURL)
}

func (s *StorageService) getLocal(fileURL string) ([]byte, error) {
	filePath := filepath.Join(s.cfg.LocalStoragePath, filepath.Base(fileURL))
	return os.ReadFile(filePath)
}

func (s *StorageService) getS3(fileURL string) ([]byte, error) {
	key := filepath.Base(fileURL)

	result, err := s.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.cfg.AWSBucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

// GenerateUniqueFilename generates a unique filename with timestamp
func GenerateUniqueFilename(prefix, extension string) string {
	timestamp := time.Now().Format("20060102-150405")
	uniqueID := uuid.New().String()[:8]
	return fmt.Sprintf("%s-%s-%s.%s", prefix, timestamp, uniqueID, extension)
}

func getContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}
