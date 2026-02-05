package main

import (
	"log"
	"os"

	"finpharm-ai/services/transaction/internal/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // default transaction port
	}

	router := httpapi.NewRouter()

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("transaction service failed to start: %v", err)
	}
}
