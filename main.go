package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"url-shortener/config"
	"url-shortener/handlers"
	"url-shortener/middleware"
	"url-shortener/storage"
)

func main() {
	cfg := config.Load()

	// Initialize storage
	store := storage.NewMemoryStorage()

	// Initialize handlers
	urlHandler := handlers.NewURLHandler(store)

	// Setup router
	r := mux.NewRouter()

	// Apply metrics middleware
	r.Use(middleware.PrometheusMiddleware)

	// Routes
	r.HandleFunc("/shorten", urlHandler.ShortenURL).Methods("POST")
	r.HandleFunc("/health", urlHandler.HealthCheck).Methods("GET")
	r.HandleFunc("/stats/{shortCode}", urlHandler.GetStats).Methods("GET")

	// Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// Access short url
	r.HandleFunc("/{shortCode}", urlHandler.RedirectURL).Methods("GET")

	// Create server
	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
