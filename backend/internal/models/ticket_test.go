package models

import (
	"testing"
	"time"
)

func TestTicketTypeIsAvailable(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		ticketType TicketType
		expected   bool
	}{
		{
			name: "Available ticket",
			ticketType: TicketType{
				IsActive:  true,
				Quantity:  100,
				Sold:      50,
				SaleStart: now.Add(-1 * time.Hour),
				SaleEnd:   now.Add(1 * time.Hour),
			},
			expected: true,
		},
		{
			name: "Inactive ticket",
			ticketType: TicketType{
				IsActive:  false,
				Quantity:  100,
				Sold:      50,
				SaleStart: now.Add(-1 * time.Hour),
				SaleEnd:   now.Add(1 * time.Hour),
			},
			expected: false,
		},
		{
			name: "Sold out ticket",
			ticketType: TicketType{
				IsActive:  true,
				Quantity:  100,
				Sold:      100,
				SaleStart: now.Add(-1 * time.Hour),
				SaleEnd:   now.Add(1 * time.Hour),
			},
			expected: false,
		},
		{
			name: "Sale not started",
			ticketType: TicketType{
				IsActive:  true,
				Quantity:  100,
				Sold:      50,
				SaleStart: now.Add(1 * time.Hour),
				SaleEnd:   now.Add(2 * time.Hour),
			},
			expected: false,
		},
		{
			name: "Sale ended",
			ticketType: TicketType{
				IsActive:  true,
				Quantity:  100,
				Sold:      50,
				SaleStart: now.Add(-2 * time.Hour),
				SaleEnd:   now.Add(-1 * time.Hour),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ticketType.IsAvailable()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTicketTypeRemainingTickets(t *testing.T) {
	tests := []struct {
		name     string
		quantity int
		sold     int
		expected int
	}{
		{"Half sold", 100, 50, 50},
		{"None sold", 100, 0, 100},
		{"All sold", 100, 100, 0},
		{"Over sold (edge case)", 100, 105, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticketType := &TicketType{
				Quantity: tt.quantity,
				Sold:     tt.sold,
			}
			result := ticketType.RemainingTickets()
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}
