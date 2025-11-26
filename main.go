package main

import (
	"forwardproxy/router"
	"log"
)

func main() {
	// Setup router
	r := router.SetupRouter()

	log.Printf("Starting server on :8081\n")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
