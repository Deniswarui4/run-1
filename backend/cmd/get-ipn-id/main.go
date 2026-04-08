package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/warui/event-ticketing-api/internal/config"
	"github.com/warui/event-ticketing-api/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Pesapal service
	pesapalService := services.NewPesapalService(cfg)

	fmt.Println("Fetching registered IPNs from Pesapal...")
	fmt.Println("==========================================")

	// List all IPNs
	ipns, err := pesapalService.ListIPNs()
	if err != nil {
		log.Fatalf("Failed to list IPNs: %v", err)
	}

	if len(ipns) == 0 {
		fmt.Println("No IPNs registered yet.")
		fmt.Println("\nRegistering IPN URL:", cfg.PesapalCallbackURL)
		
		// Register the IPN
		ipn, err := pesapalService.RegisterIPN(cfg.PesapalCallbackURL, "GET")
		if err != nil {
			log.Fatalf("Failed to register IPN: %v", err)
		}

		fmt.Println("\n✓ IPN registered successfully!")
		fmt.Printf("IPN ID: %s\n", ipn.IPNID)
		fmt.Printf("URL: %s\n", ipn.URL)
		fmt.Printf("\nAdd this to your .env file:\n")
		fmt.Printf("PESAPAL_IPN_ID=%s\n", ipn.IPNID)
		return
	}

	fmt.Printf("Found %d registered IPN(s):\n\n", len(ipns))
	for i, ipn := range ipns {
		fmt.Printf("%d. IPN ID: %s\n", i+1, ipn.IPNID)
		fmt.Printf("   URL: %s\n", ipn.URL)
		fmt.Printf("   Type: %s\n", ipn.IPNNotificationType)
		fmt.Printf("   Created: %s\n", ipn.CreatedDate)
		fmt.Println()
	}

	// Find the matching IPN
	for _, ipn := range ipns {
		if ipn.URL == cfg.PesapalCallbackURL {
			fmt.Printf("✓ Found matching IPN for %s\n", cfg.PesapalCallbackURL)
			fmt.Printf("\nAdd this to your .env file:\n")
			fmt.Printf("PESAPAL_IPN_ID=%s\n", ipn.IPNID)
			return
		}
	}

	fmt.Printf("⚠ No IPN found for %s\n", cfg.PesapalCallbackURL)
	fmt.Println("\nRegistering new IPN...")
	
	ipn, err := pesapalService.RegisterIPN(cfg.PesapalCallbackURL, "GET")
	if err != nil {
		log.Fatalf("Failed to register IPN: %v", err)
	}

	fmt.Println("\n✓ IPN registered successfully!")
	fmt.Printf("IPN ID: %s\n", ipn.IPNID)
	fmt.Printf("\nAdd this to your .env file:\n")
	fmt.Printf("PESAPAL_IPN_ID=%s\n", ipn.IPNID)
}
