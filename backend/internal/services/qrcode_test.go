package services

import (
	"testing"
)

func TestNewQRCodeService(t *testing.T) {
	service := NewQRCodeService()
	if service == nil {
		t.Error("QRCodeService should not be nil")
	}
}

func TestGenerateQRCode(t *testing.T) {
	service := NewQRCodeService()
	data := "TEST-DATA-12345"

	qrCode, err := service.GenerateQRCode(data, 256)
	if err != nil {
		t.Fatalf("GenerateQRCode failed: %v", err)
	}

	if len(qrCode) == 0 {
		t.Error("QR code data should not be empty")
	}

	// PNG files start with specific bytes
	if qrCode[0] != 0x89 || qrCode[1] != 0x50 || qrCode[2] != 0x4E || qrCode[3] != 0x47 {
		t.Error("Generated data should be a valid PNG file")
	}
}

func TestGenerateQRCodeDefaultSize(t *testing.T) {
	service := NewQRCodeService()
	data := "TEST-DATA"

	qrCode, err := service.GenerateQRCode(data, 0) // 0 should use default size
	if err != nil {
		t.Fatalf("GenerateQRCode with default size failed: %v", err)
	}

	if len(qrCode) == 0 {
		t.Error("QR code data should not be empty")
	}
}

func TestGenerateTicketQRCode(t *testing.T) {
	service := NewQRCodeService()
	ticketNumber := "TKT-12345678"
	ticketID := "uuid-test-id"

	qrCode, err := service.GenerateTicketQRCode(ticketNumber, ticketID)
	if err != nil {
		t.Fatalf("GenerateTicketQRCode failed: %v", err)
	}

	if len(qrCode) == 0 {
		t.Error("QR code data should not be empty")
	}
}

func TestGenerateQRCodeWithDifferentData(t *testing.T) {
	service := NewQRCodeService()

	qrCode1, _ := service.GenerateQRCode("DATA1", 256)
	qrCode2, _ := service.GenerateQRCode("DATA2", 256)

	if len(qrCode1) == len(qrCode2) {
		// QR codes with different data might have different sizes
		// This is not always true, but we can check they're not identical
		identical := true
		for i := range qrCode1 {
			if qrCode1[i] != qrCode2[i] {
				identical = false
				break
			}
		}
		if identical {
			t.Error("QR codes with different data should not be identical")
		}
	}
}
