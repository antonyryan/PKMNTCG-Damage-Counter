package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pkmntcg/backend/internal/app"
)

// main wires the HTTP server to the configured Gin router and handles graceful shutdown.
func main() {
	application := app.NewFromEnv()

	// Start server in a background goroutine so we can listen for shutdown signals.
	go func() {
		log.Printf("server listening on %s", application.Server.Addr)
		if err := application.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutdown signal received, draining...")

	// Give in-flight HTTP requests up to 10 s to finish.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := application.Server.Shutdown(ctx); err != nil {
		log.Printf("server shutdown: %v", err)
	}

	// Drain the analytics worker and close the SQLite connection.
	application.Analytics.Shutdown()
	log.Println("shutdown complete")
}
