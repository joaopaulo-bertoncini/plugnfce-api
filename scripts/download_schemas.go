package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/validator"
)

func main() {
	ctx := context.Background()

	// Create validator with schemas directory
	v, err := validator.NewXMLValidator("./internal/infrastructure/sefaz/schemas")
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	fmt.Println("Downloading SEFAZ schemas for NFC-e v4.00...")

	// Download schemas
	if err := v.DownloadSEFAZSchemas(ctx, "4.00"); err != nil {
		log.Fatalf("Failed to download schemas: %v", err)
	}

	fmt.Println("Schemas downloaded successfully!")

	// List available schemas
	schemas, err := v.ListAvailableSchemas()
	if err != nil {
		log.Fatalf("Failed to list schemas: %v", err)
	}

	fmt.Println("Available schemas:")
	for _, schema := range schemas {
		fmt.Printf("  - %s\n", schema)
	}
}
