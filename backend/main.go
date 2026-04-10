package main

// main wires the HTTP server to the configured Gin router.
func main() {
	r := setupRouter()
	r.Run(":8080")
}
