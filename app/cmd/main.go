package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ashwinkg/monitoring-service/internal/config"
	"github.com/ashwinkg/monitoring-service/internal/repository"
)

func main() {
	cfg := config.Load()

	// Connect to dependencies (Postgres, Kafka, Redis)
	repository.NewPostgres(cfg)
	repository.NewRedis(cfg)

	// Graceful shutdown and other server setup can be added here
	// go func() {
	// 	log.Printf("Server listening on %s", cfg.AppPort)
	// 	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	//     log.Fatalf("Server error: %v", err)
	// }
	// }()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	_, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	log.Println("Server exited cleanly")
}
