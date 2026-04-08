package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/warui/event-ticketing-api/internal/models"
)

func TestNewPDFService(t *testing.T) {
	service := NewPDFService()
	if service == nil {
		t.Error("PDFService should not be nil")
	}
}

func TestGenerateTicketPDF(t *testing.T) {
	service := NewPDFService()

	// Create test data
	event := &models.Event{
		ID:          uuid.New(),
		Title:       "Test Concert",
		Description: "A great test concert",
		Category:    "Music",
		Venue:       "Test Arena",
		Address:     "123 Test Street",
		City:        "Test City",
		Country:     "Test Country",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(28 * time.Hour),
	}

	attendee := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	ticket := &models.Ticket{
		ID:           uuid.New(),
		TicketNumber: "TKT-12345678",
		Price:        5000,
		Status:       models.TicketStatusConfirmed,
		TicketType: models.TicketType{
			Name:        "General Admission",
			Description: "Standard entry",
		},
	}

	// Generate simple QR code data
	qrService := NewQRCodeService()
	qrData, _ := qrService.GenerateTicketQRCode(ticket.TicketNumber, ticket.ID.String())

	// Generate PDF
	pdfData, err := service.GenerateTicketPDF(ticket, event, attendee, qrData)
	if err != nil {
		t.Fatalf("GenerateTicketPDF failed: %v", err)
	}

	if len(pdfData) == 0 {
		t.Error("PDF data should not be empty")
	}

	// PDF files start with %PDF
	if pdfData[0] != '%' || pdfData[1] != 'P' || pdfData[2] != 'D' || pdfData[3] != 'F' {
		t.Error("Generated data should be a valid PDF file")
	}
}

func TestGenerateTicketPDFWithoutQRCode(t *testing.T) {
	service := NewPDFService()

	event := &models.Event{
		ID:        uuid.New(),
		Title:     "Test Event",
		Venue:     "Test Venue",
		StartDate: time.Now().Add(24 * time.Hour),
		EndDate:   time.Now().Add(28 * time.Hour),
	}

	attendee := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	ticket := &models.Ticket{
		ID:           uuid.New(),
		TicketNumber: "TKT-87654321",
		Price:        3000,
		Status:       models.TicketStatusConfirmed,
	}

	// Generate PDF without QR code
	pdfData, err := service.GenerateTicketPDF(ticket, event, attendee, nil)
	if err != nil {
		t.Fatalf("GenerateTicketPDF without QR code failed: %v", err)
	}

	if len(pdfData) == 0 {
		t.Error("PDF data should not be empty")
	}
}
