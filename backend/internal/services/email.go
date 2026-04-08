package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/resendlabs/resend-go/v2"
	"github.com/warui/event-ticketing-api/internal/config"
	"github.com/warui/event-ticketing-api/internal/models"
)

type EmailService struct {
	client *resend.Client
	cfg    *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	client := resend.NewClient(cfg.ResendAPIKey)
	return &EmailService{
		client: client,
		cfg:    cfg,
	}
}

// SendWelcomeEmail sends welcome email to new users
func (e *EmailService) SendWelcomeEmail(user *models.User) error {
	if e.cfg.ResendAPIKey == "" {
		return fmt.Errorf("email service not configured")
	}

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", e.cfg.FromName, e.cfg.FromEmail),
		To:      []string{user.Email},
		Subject: "Welcome to Event Ticketing Platform",
		Html: fmt.Sprintf(`
			<h1>Welcome, %s!</h1>
			<p>Thank you for joining our Event Ticketing Platform.</p>
			<p>Your account has been created successfully with the role: <strong>%s</strong></p>
			<p>Start exploring amazing events and book your tickets today!</p>
			<br>
			<p>Best regards,<br>Event Ticketing Team</p>
		`, user.FirstName, user.Role),
	}

	_, err := e.client.Emails.Send(params)
	return err
}

// SendTicketEmail sends ticket confirmation email with PDF attachment
func (e *EmailService) SendTicketEmail(ticket *models.Ticket, event *models.Event, user *models.User, pdfData []byte) error {
	if e.cfg.ResendAPIKey == "" {
		return fmt.Errorf("email service not configured")
	}

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", e.cfg.FromName, e.cfg.FromEmail),
		To:      []string{user.Email},
		Subject: fmt.Sprintf("Your Ticket for %s", event.Title),
		Html: fmt.Sprintf(`
			<h1>Ticket Confirmation</h1>
			<p>Hi %s,</p>
			<p>Your ticket has been confirmed for:</p>
			<h2>%s</h2>
			<p><strong>Ticket Number:</strong> %s</p>
			<p><strong>Venue:</strong> %s</p>
			<p><strong>Date:</strong> %s</p>
			<p><strong>Price:</strong> %s %.2f</p>
			<br>
			<p>Please find your ticket PDF attached. Show the QR code at the venue for entry.</p>
			<p>See you at the event!</p>
			<br>
			<p>Best regards,<br>Event Ticketing Team</p>
		`, user.FirstName, event.Title, ticket.TicketNumber, event.Venue,
			event.StartDate.Format("Mon, Jan 2, 2006 at 3:04 PM"), e.cfg.Currency, ticket.Price),
	}

	// Attach PDF if available
	if pdfData != nil && len(pdfData) > 0 {
		params.Attachments = []*resend.Attachment{
			{
				Filename: fmt.Sprintf("ticket-%s.pdf", ticket.TicketNumber),
				Content:  pdfData,
			},
		}
	}

	_, err := e.client.Emails.Send(params)
	return err
}

// SendEventApprovalEmail notifies organizer about event approval
func (e *EmailService) SendEventApprovalEmail(event *models.Event, organizer *models.User, approved bool) error {
	if e.cfg.ResendAPIKey == "" {
		return fmt.Errorf("email service not configured")
	}

	status := "Approved"
	message := "Your event has been approved and is now live!"
	if !approved {
		status = "Rejected"
		message = fmt.Sprintf("Your event has been rejected. Reason: %s", event.ModerationComment)
	}

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", e.cfg.FromName, e.cfg.FromEmail),
		To:      []string{organizer.Email},
		Subject: fmt.Sprintf("Event %s: %s", status, event.Title),
		Html: fmt.Sprintf(`
			<h1>Event %s</h1>
			<p>Hi %s,</p>
			<p>%s</p>
			<h3>Event Details:</h3>
			<p><strong>Title:</strong> %s</p>
			<p><strong>Date:</strong> %s</p>
			<p><strong>Venue:</strong> %s</p>
			<br>
			<p>Best regards,<br>Event Ticketing Team</p>
		`, status, organizer.FirstName, message, event.Title,
			event.StartDate.Format("Mon, Jan 2, 2006"), event.Venue),
	}

	_, err := e.client.Emails.Send(params)
	return err
}

// SendWithdrawalStatusEmail notifies organizer about withdrawal request status
func (e *EmailService) SendWithdrawalStatusEmail(withdrawal *models.WithdrawalRequest, organizer *models.User) error {
	if e.cfg.ResendAPIKey == "" {
		return fmt.Errorf("email service not configured")
	}

	var statusMessage string
	switch withdrawal.Status {
	case models.WithdrawalStatusApproved:
		statusMessage = "Your withdrawal request has been approved and is being processed."
	case models.WithdrawalStatusRejected:
		statusMessage = fmt.Sprintf("Your withdrawal request has been rejected. Reason: %s", withdrawal.ReviewComment)
	case models.WithdrawalStatusProcessed:
		statusMessage = "Your withdrawal has been processed successfully. Funds should arrive in your account shortly."
	}

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", e.cfg.FromName, e.cfg.FromEmail),
		To:      []string{organizer.Email},
		Subject: fmt.Sprintf("Withdrawal Request %s", withdrawal.Status),
		Html: fmt.Sprintf(`
			<h1>Withdrawal Request Update</h1>
			<p>Hi %s,</p>
			<p>%s</p>
			<h3>Withdrawal Details:</h3>
			<p><strong>Amount:</strong> %s %.2f</p>
			<p><strong>Fee:</strong> %s %.2f</p>
			<p><strong>Net Amount:</strong> %s %.2f</p>
			<p><strong>Bank:</strong> %s</p>
			<p><strong>Account:</strong> %s</p>
			<br>
			<p>Best regards,<br>Event Ticketing Team</p>
		`, organizer.FirstName, statusMessage, e.cfg.Currency, withdrawal.Amount,
			e.cfg.Currency, withdrawal.WithdrawalFee, e.cfg.Currency, withdrawal.NetAmount,
			withdrawal.BankName, withdrawal.AccountNumber),
	}

	_, err := e.client.Emails.Send(params)
	return err
}

// GenerateVerificationToken generates a random verification token
func (e *EmailService) GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SendVerificationEmail sends email verification link
func (e *EmailService) SendVerificationEmail(user *models.User, token string) error {
	if e.cfg.ResendAPIKey == "" {
		return fmt.Errorf("email service not configured")
	}

	verificationURL := fmt.Sprintf("%s/verify?token=%s", e.cfg.FrontendURL, token)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", e.cfg.FromName, e.cfg.FromEmail),
		To:      []string{user.Email},
		Subject: "Verify Your Email Address",
		Html: fmt.Sprintf(`
			<h1>Welcome to Event Ticketing Platform!</h1>
			<p>Hi %s,</p>
			<p>Thank you for registering with us. To complete your registration, please verify your email address by clicking the button below:</p>
			<br>
			<div style="text-align: center; margin: 30px 0;">
				<a href="%s" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; display: inline-block; font-weight: bold;">
					Verify Email Address
				</a>
			</div>
			<br>
			<p>If the button doesn't work, you can also copy and paste this link into your browser:</p>
			<p><a href="%s">%s</a></p>
			<br>
			<p><strong>This link will expire in 24 hours.</strong></p>
			<br>
			<p>If you didn't create an account with us, please ignore this email.</p>
			<br>
			<p>Best regards,<br>Event Ticketing Team</p>
		`, user.FirstName, verificationURL, verificationURL, verificationURL),
	}

	_, err := e.client.Emails.Send(params)
	return err
}

// SendPasswordResetEmail sends password reset link
func (e *EmailService) SendPasswordResetEmail(user *models.User, token string) error {
	if e.cfg.ResendAPIKey == "" {
		return fmt.Errorf("email service not configured")
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", e.cfg.FrontendURL, token)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", e.cfg.FromName, e.cfg.FromEmail),
		To:      []string{user.Email},
		Subject: "Reset Your Password",
		Html: fmt.Sprintf(`
			<h1>Password Reset Request</h1>
			<p>Hi %s,</p>
			<p>We received a request to reset your password. Click the button below to create a new password:</p>
			<br>
			<div style="text-align: center; margin: 30px 0;">
				<a href="%s" style="background-color: #dc3545; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; display: inline-block; font-weight: bold;">
					Reset Password
				</a>
			</div>
			<br>
			<p>If the button doesn't work, you can also copy and paste this link into your browser:</p>
			<p><a href="%s">%s</a></p>
			<br>
			<p><strong>This link will expire in 1 hour.</strong></p>
			<br>
			<p>If you didn't request a password reset, please ignore this email. Your password will remain unchanged.</p>
			<br>
			<p>Best regards,<br>Event Ticketing Team</p>
		`, user.FirstName, resetURL, resetURL, resetURL),
	}

	_, err := e.client.Emails.Send(params)
	return err
}
