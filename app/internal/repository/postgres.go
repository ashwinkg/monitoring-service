package repository

import (
	"log"
	"time"

	"github.com/ashwinkg/monitoring-service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Event is persisted to Postgres.
type Event struct {
	ID        uint `gorm:"primaryKey"`
	Type      string
	Payload   string
	CreatedAt time.Time
}

func NewPostgres(cfg *config.Config) *gorm.DB {
	var db *gorm.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
		if err == nil {
			break
		}

		log.Printf("Waiting for Postgres to be ready... (%d/10)", i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}

	if err := db.AutoMigrate(&Event{}); err != nil {
		log.Fatalf("Failed to migrate Postgres schema: %v", err)
	}

	log.Println("Connected to Postgres")
	return db
}
