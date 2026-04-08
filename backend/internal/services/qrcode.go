package services

import (
	"fmt"

	"github.com/skip2/go-qrcode"
)

type QRCodeService struct{}

func NewQRCodeService() *QRCodeService {
	return &QRCodeService{}
}

// GenerateQRCode generates a QR code for a ticket
func (q *QRCodeService) GenerateQRCode(data string, size int) ([]byte, error) {
	if size == 0 {
		size = 256
	}

	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	qr.DisableBorder = false

	return qr.PNG(size)
}

// GenerateTicketQRCode generates QR code specifically for tickets
func (q *QRCodeService) GenerateTicketQRCode(ticketNumber, ticketID string) ([]byte, error) {
	// Create QR code data with ticket information
	data := fmt.Sprintf("TICKET:%s:ID:%s", ticketNumber, ticketID)
	return q.GenerateQRCode(data, 256)
}
