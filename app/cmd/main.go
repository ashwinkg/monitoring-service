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
	pg := repository.NewPostgres(cfg)
	rdb := repository.NewRedis(cfg)
	kafka := repository.NewKafka(cfg)

	//start consuming messages
	kafka.Consume(context.Background(), func(key, value string) {
		log.Printf("Received message - Key: %s, Value: %s", key, value)
	})

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	_, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Close all connections
	kafka.Close()
	rdb.Close()
	pg.Close()

	log.Println("Server exited cleanly")
}
