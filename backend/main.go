package main

import (
	"log"
	"os"
)

// main wires the HTTP server to the configured Gin router.
func main() {
	r := setupRouter()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
