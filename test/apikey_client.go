package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Samuel-Ijegbulem/go-pingdom/pingdom"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("PINGDOM_API_KEY")
	if apiKey == "" {
		log.Fatal("PINGDOM_API_KEY environment variable is not set")
	}

	fmt.Println("Testing Pingdom API Key Authentication...")

	// Create client with API Key authentication
	client, err := pingdom.NewClientWithConfig(pingdom.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Error creating Pingdom client: %v", err)
	}

	// Test the connection by getting a list of checks
	fmt.Println("Fetching checks...")
	checks, err := client.Checks.List()
	if err != nil {
		log.Fatalf("Error listing checks: %v", err)
	}

	fmt.Printf("Success! Found %d checks:\n", len(checks))
	for i, check := range checks {
		if i >= 5 {
			fmt.Println("... (more checks not shown)")
			break
		}
		fmt.Printf("- Check ID: %d, Name: %s, Status: %s\n", check.ID, check.Name, check.Status)
	}

	fmt.Println("\nAPI Key authentication test completed successfully!")
}
