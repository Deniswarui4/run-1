package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	baseURL := getEnv("BACKEND_URL", "http://localhost:8080")

	// Test 1: Get a published event
	fmt.Println("=== Test 1: Fetching published events ===")
	resp, err := http.Get(baseURL + "/api/v1/events")
	if err != nil {
		fmt.Printf("Error fetching events: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	
	var eventsResp struct {
		Events []struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"events"`
	}
	
	if err := json.Unmarshal(body, &eventsResp); err != nil {
		fmt.Printf("Error parsing events: %v\n", err)
		fmt.Printf("Response: %s\n", string(body))
		return
	}

	if len(eventsResp.Events) == 0 {
		fmt.Println("No events found. Please create a published event first.")
		return
	}

	event := eventsResp.Events[0]
	fmt.Printf("Found event: %s (ID: %s, Status: %s)\n\n", event.Title, event.ID, event.Status)

	// Test 2: Get event details with ticket types
	fmt.Println("=== Test 2: Fetching event details ===")
	resp, err = http.Get(baseURL + "/api/v1/events/" + event.ID)
	if err != nil {
		fmt.Printf("Error fetching event details: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)

	var eventDetail struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		TicketTypes []struct {
			ID          string  `json:"id"`
			Name        string  `json:"name"`
			Price       float64 `json:"price"`
			Quantity    int     `json:"quantity"`
			Sold        int     `json:"sold"`
			MaxPerOrder int     `json:"max_per_order"`
			IsActive    bool    `json:"is_active"`
		} `json:"ticket_types"`
	}

	if err := json.Unmarshal(body, &eventDetail); err != nil {
		fmt.Printf("Error parsing event details: %v\n", err)
		fmt.Printf("Response: %s\n", string(body))
		return
	}

	if len(eventDetail.TicketTypes) == 0 {
		fmt.Println("No ticket types found for this event. Please create ticket types first.")
		return
	}

	ticketType := eventDetail.TicketTypes[0]
	fmt.Printf("Found ticket type: %s (ID: %s, Price: %.2f, Available: %d)\n\n",
		ticketType.Name, ticketType.ID, ticketType.Price, ticketType.Quantity-ticketType.Sold)

	// Test 3: Attempt guest checkout
	fmt.Println("=== Test 3: Testing guest checkout ===")
	checkoutReq := map[string]interface{}{
		"email":    "test-guest@example.com",
		"event_id": event.ID,
		"items": []map[string]interface{}{
			{
				"ticket_type_id": ticketType.ID,
				"quantity":       1,
			},
		},
	}

	reqBody, _ := json.Marshal(checkoutReq)
	fmt.Printf("Request: %s\n", string(reqBody))

	resp, err = http.Post(
		baseURL+"/api/v1/guest/checkout",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		fmt.Printf("Error initiating guest checkout: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n\n", string(body))

	if resp.StatusCode != 200 {
		fmt.Println("❌ Guest checkout failed!")
		fmt.Println("\nPossible issues:")
		fmt.Println("1. Check if the event is published")
		fmt.Println("2. Check if ticket types are active and available")
		fmt.Println("3. Check database schema for guest_email column")
		fmt.Println("4. Check Paystack configuration")
		fmt.Println("5. Review backend logs for detailed error messages")
	} else {
		fmt.Println("✓ Guest checkout initiated successfully!")
		
		var checkoutResp map[string]interface{}
		json.Unmarshal(body, &checkoutResp)
		if authURL, ok := checkoutResp["authorization_url"].(string); ok {
			fmt.Printf("\nPayment URL: %s\n", authURL)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
