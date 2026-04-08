package services

import (
	"fmt"

	"github.com/h2non/bimg"
)

type ImageService struct{}

func NewImageService() *ImageService {
	return &ImageService{}
}

// ProcessEventImage processes and optimizes event images
func (i *ImageService) ProcessEventImage(imageData []byte, maxWidth, maxHeight int) ([]byte, error) {
	if maxWidth == 0 {
		maxWidth = 1200
	}
	if maxHeight == 0 {
		maxHeight = 800
	}

	image := bimg.NewImage(imageData)

	// Get image metadata
	size, err := image.Size()
	if err != nil {
		return nil, fmt.Errorf("failed to get image size: %w", err)
	}

	// Calculate new dimensions maintaining aspect ratio
	newWidth := size.Width
	newHeight := size.Height

	if size.Width > maxWidth {
		newWidth = maxWidth
		newHeight = (size.Height * maxWidth) / size.Width
	}

	if newHeight > maxHeight {
		newHeight = maxHeight
		newWidth = (size.Width * maxHeight) / size.Height
	}

	// Process image
	options := bimg.Options{
		Width:   newWidth,
		Height:  newHeight,
		Quality: 85,
		Crop:    false,
		Type:    bimg.JPEG,
	}

	processed, err := image.Process(options)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	return processed, nil
}

// CreateThumbnail creates a thumbnail version of an image
func (i *ImageService) CreateThumbnail(imageData []byte, size int) ([]byte, error) {
	if size == 0 {
		size = 300
	}

	image := bimg.NewImage(imageData)

	options := bimg.Options{
		Width:   size,
		Height:  size,
		Quality: 80,
		Crop:    true,
		Gravity: bimg.GravityCentre,
		Type:    bimg.JPEG,
	}

	thumbnail, err := image.Process(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create thumbnail: %w", err)
	}

	return thumbnail, nil
}

// ValidateImage checks if the data is a valid image
func (i *ImageService) ValidateImage(imageData []byte) error {
	image := bimg.NewImage(imageData)

	imageType := bimg.DetermineImageType(imageData)
	if imageType == bimg.UNKNOWN {
		return fmt.Errorf("unknown or invalid image format")
	}

	// Check if image can be read
	_, err := image.Size()
	if err != nil {
		return fmt.Errorf("invalid image data: %w", err)
	}

	return nil
}
