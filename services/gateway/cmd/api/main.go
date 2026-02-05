package main

import (
	"log"
	"os"

	"finpharm-ai/services/gateway/internal/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default gateway port
	}

	router := httpapi.NewRouter()

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("gateway failed to start: %v", err)
	}
}
