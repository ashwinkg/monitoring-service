package main

import (
	"github.com/ashwinkg/monitoring-service/internal/config"
	"github.com/ashwinkg/monitoring-service/internal/repository"
)

func main() {
	cfg := config.Load()

	// Connect to dependencies (Postgres, Kafka, Redis)
	repository.NewPostgres(cfg)
}
