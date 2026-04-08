package services

import (
	"testing"

	"github.com/warui/event-ticketing-api/internal/config"
)

func TestGenerateUniqueFilename(t *testing.T) {
	prefix := "test"
	extension := "jpg"

	filename1 := GenerateUniqueFilename(prefix, extension)
	filename2 := GenerateUniqueFilename(prefix, extension)

	if filename1 == "" {
		t.Error("Filename should not be empty")
	}

	if filename1 == filename2 {
		t.Error("Generated filenames should be unique")
	}

	// Check format
	if len(filename1) < 20 {
		t.Error("Filename should contain prefix, timestamp, and UUID")
	}
}

func TestNewStorageService(t *testing.T) {
	cfg := &config.Config{
		StorageType:      "local",
		LocalStoragePath: "./test_storage",
	}

	service, err := NewStorageService(cfg)
	if err != nil {
		t.Fatalf("NewStorageService failed: %v", err)
	}

	if service == nil {
		t.Error("Service should not be nil")
	}

	if !service.useLocal {
		t.Error("Service should use local storage")
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"image.jpg", "image/jpeg"},
		{"image.jpeg", "image/jpeg"},
		{"image.png", "image/png"},
		{"image.gif", "image/gif"},
		{"document.pdf", "application/pdf"},
		{"data.json", "application/json"},
		{"unknown.xyz", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := getContentType(tt.filename)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
