package services

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/warui/event-ticketing-api/internal/models"
)

type PDFService struct{}

func NewPDFService() *PDFService {
	return &PDFService{}
}

// GenerateTicketPDF generates a PDF ticket
func (p *PDFService) GenerateTicketPDF(ticket *models.Ticket, event *models.Event, attendee *models.User, qrCodeData []byte) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set colors
	primaryColor := func() { pdf.SetTextColor(41, 128, 185) }
	blackColor := func() { pdf.SetTextColor(0, 0, 0) }
	grayColor := func() { pdf.SetTextColor(128, 128, 128) }

	// Header
	pdf.SetFont("Arial", "B", 24)
	primaryColor()
	pdf.Cell(0, 15, "EVENT TICKET")
	pdf.Ln(20)

	// Event Title
	pdf.SetFont("Arial", "B", 18)
	blackColor()
	pdf.MultiCell(0, 10, event.Title, "", "L", false)
	pdf.Ln(5)

	// Ticket Number
	pdf.SetFont("Arial", "B", 14)
	primaryColor()
	pdf.Cell(0, 10, fmt.Sprintf("Ticket #%s", ticket.TicketNumber))
	pdf.Ln(15)

	// Attendee Information
	pdf.SetFont("Arial", "B", 12)
	blackColor()
	pdf.Cell(0, 8, "Attendee Information")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	grayColor()
	pdf.Cell(50, 7, "Name:")
	blackColor()
	pdf.Cell(0, 7, fmt.Sprintf("%s %s", attendee.FirstName, attendee.LastName))
	pdf.Ln(7)

	grayColor()
	pdf.Cell(50, 7, "Email:")
	blackColor()
	pdf.Cell(0, 7, attendee.Email)
	pdf.Ln(12)

	// Event Details
	pdf.SetFont("Arial", "B", 12)
	blackColor()
	pdf.Cell(0, 8, "Event Details")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	grayColor()
	pdf.Cell(50, 7, "Venue:")
	blackColor()
	pdf.MultiCell(0, 7, event.Venue, "", "L", false)

	if event.Address != "" {
		grayColor()
		pdf.Cell(50, 7, "Address:")
		blackColor()
		pdf.MultiCell(0, 7, event.Address, "", "L", false)
	}

	grayColor()
	pdf.Cell(50, 7, "Date & Time:")
	blackColor()
	pdf.Cell(0, 7, event.StartDate.Format("Monday, January 2, 2006 at 3:04 PM"))
	pdf.Ln(7)

	if !event.EndDate.IsZero() && !event.EndDate.Equal(event.StartDate) {
		grayColor()
		pdf.Cell(50, 7, "End Date:")
		blackColor()
		pdf.Cell(0, 7, event.EndDate.Format("Monday, January 2, 2006 at 3:04 PM"))
		pdf.Ln(7)
	}

	grayColor()
	pdf.Cell(50, 7, "Category:")
	blackColor()
	pdf.Cell(0, 7, event.Category)
	pdf.Ln(12)

	// Ticket Type and Price
	if ticket.TicketType.Name != "" {
		pdf.SetFont("Arial", "B", 12)
		blackColor()
		pdf.Cell(0, 8, "Ticket Information")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 11)
		grayColor()
		pdf.Cell(50, 7, "Ticket Type:")
		blackColor()
		pdf.Cell(0, 7, ticket.TicketType.Name)
		pdf.Ln(7)
	}

	grayColor()
	pdf.Cell(50, 7, "Price:")
	blackColor()
	pdf.Cell(0, 7, fmt.Sprintf("NGN %.2f", ticket.Price))
	pdf.Ln(7)

	grayColor()
	pdf.Cell(50, 7, "Status:")
	blackColor()
	pdf.Cell(0, 7, string(ticket.Status))
	pdf.Ln(15)

	// QR Code
	if qrCodeData != nil {
		pdf.SetFont("Arial", "B", 12)
		blackColor()
		pdf.Cell(0, 8, "Entry QR Code")
		pdf.Ln(10)

		// Register QR code image
		reader := bytes.NewReader(qrCodeData)
		imageInfo := pdf.RegisterImageReader(fmt.Sprintf("qr_%s", ticket.ID.String()), "PNG", reader)
		if imageInfo != nil {
			// Center the QR code
			pageWidth, _ := pdf.GetPageSize()
			qrSize := 60.0
			xPos := (pageWidth - qrSize) / 2
			pdf.Image(fmt.Sprintf("qr_%s", ticket.ID.String()), xPos, pdf.GetY(), qrSize, qrSize, false, "", 0, "")
			pdf.Ln(65)
		}

		pdf.SetFont("Arial", "I", 9)
		grayColor()
		pdf.Cell(0, 5, "Please present this QR code at the venue entrance")
		pdf.Ln(10)
	}

	// Footer
	pdf.SetY(-30)
	pdf.SetFont("Arial", "I", 8)
	grayColor()
	pdf.Cell(0, 5, fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006 at 3:04 PM")))
	pdf.Ln(5)
	pdf.Cell(0, 5, "This ticket is non-transferable and valid for single entry only")
	pdf.Ln(5)
	pdf.Cell(0, 5, "For support, contact: support@eventtickets.com")

	// Output PDF to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}
