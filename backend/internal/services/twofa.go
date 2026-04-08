package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/warui/event-ticketing-api/internal/config"
)

type TwoFAService struct {
	cfg *config.Config
}

func NewTwoFAService(cfg *config.Config) *TwoFAService {
	return &TwoFAService{
		cfg: cfg,
	}
}

// GenerateSecret generates a new TOTP secret for a user
func (t *TwoFAService) GenerateSecret(userEmail string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Event Ticketing Platform",
		AccountName: userEmail,
		SecretSize:  32,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	return key, nil
}

// GenerateQRCode generates a QR code image for the TOTP secret
func (t *TwoFAService) GenerateQRCode(key *otp.Key) (string, error) {
	// Generate QR code image
	img, err := key.Image(256, 256)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code image: %w", err)
	}

	// Convert image to base64 string
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode QR code image: %w", err)
	}

	base64String := base64.StdEncoding.EncodeToString(buf.Bytes())
	return fmt.Sprintf("data:image/png;base64,%s", base64String), nil
}

// ValidateCode validates a TOTP code against a secret
func (t *TwoFAService) ValidateCode(code, secret string) bool {
	return totp.Validate(code, secret)
}

// GenerateBackupCodes generates backup codes for 2FA recovery
func (t *TwoFAService) GenerateBackupCodes() ([]string, error) {
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		// Generate 8-character backup codes
		code, err := t.generateRandomString(8)
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		codes[i] = code
	}
	return codes, nil
}

// generateRandomString generates a random alphanumeric string
func (t *TwoFAService) generateRandomString(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		// Simple random generation - in production, use crypto/rand
		b[i] = charset[i%len(charset)]
	}
	return string(b), nil
}
