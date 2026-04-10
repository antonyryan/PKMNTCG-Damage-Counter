package main

import "log"

// main wires the HTTP server to the configured Gin router.
func main() {
	r := setupRouter()
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
