// Feature: guest-checkout-cart-sharing-metrics
package services

import "net/mail"

// validateGuestEmail returns true if the email is a valid RFC 5322 address.
func validateGuestEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// ValidateGuestEmail is the exported version for use in handlers and tests.
func ValidateGuestEmail(email string) bool {
	return validateGuestEmail(email)
}
