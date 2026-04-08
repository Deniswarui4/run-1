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
	guestSessionID := "test-guest-session-12345678"

	// Test 1: Get a published event with ticket types
	fmt.Println("=== Test 1: Fetching published event ===")
	resp, err := http.Get(baseURL + "/api/v1/events")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("\n❌ Backend server is not running!")
		fmt.Println("Start it with: go run ./cmd/api")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	var eventsResp struct {
		Events []struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
		} `json:"events"`
	}
	
	if err := json.Unmarshal(body, &eventsResp); err != nil || len(eventsResp.Events) == 0 {
		fmt.Println("No events found. Please create a published event first.")
		return
	}

	eventID := eventsResp.Events[0].ID
	fmt.Printf("Using event: %s (ID: %s)\n\n", eventsResp.Events[0].Title, eventID)

	// Test 2: Get event details to find ticket types
	fmt.Println("=== Test 2: Fetching event details ===")
	resp, err = http.Get(baseURL + "/api/v1/events/" + eventID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)

	var eventDetail struct {
		TicketTypes []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"ticket_types"`
	}

	if err := json.Unmarshal(body, &eventDetail); err != nil || len(eventDetail.TicketTypes) == 0 {
		fmt.Println("No ticket types found. Please create ticket types first.")
		return
	}

	ticketTypeID := eventDetail.TicketTypes[0].ID
	fmt.Printf("Using ticket type: %s (ID: %s)\n\n", eventDetail.TicketTypes[0].Name, ticketTypeID)

	// Test 3: Try to save cart with guest session
	fmt.Println("=== Test 3: Testing cart upsert with guest session ===")
	cartReq := map[string]interface{}{
		"event_id": eventID,
		"items": []map[string]interface{}{
			{
				"ticket_type_id": ticketTypeID,
				"quantity":       2,
			},
		},
	}

	reqBody, _ := json.Marshal(cartReq)
	fmt.Printf("Request body: %s\n", string(reqBody))
	fmt.Printf("Guest session ID: %s\n", guestSessionID)

	req, _ := http.NewRequest("PUT", baseURL+"/api/v1/cart", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Guest-Session", guestSessionID)

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n\n", string(body))

	if resp.StatusCode != 200 {
		fmt.Println("❌ Cart upsert failed!")
		fmt.Println("\nPossible issues:")
		fmt.Println("1. Check if X-Guest-Session header is being received")
		fmt.Println("2. Check backend logs for detailed error")
		fmt.Println("3. Verify draft_carts table exists")
		return
	}

	fmt.Println("✓ Cart saved successfully!")

	// Test 4: Retrieve the cart
	fmt.Println("\n=== Test 4: Retrieving cart ===")
	req, _ = http.NewRequest("GET", baseURL+"/api/v1/cart?event_id="+eventID, nil)
	req.Header.Set("X-Guest-Session", guestSessionID)

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n\n", string(body))

	if resp.StatusCode == 200 {
		fmt.Println("✓ Cart retrieved successfully!")
	} else {
		fmt.Println("❌ Cart retrieval failed!")
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
