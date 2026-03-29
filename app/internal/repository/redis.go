package repository

import (
	"context"
	"log"

	"github.com/ashwinkg/monitoring-service/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedis(cfg *config.Config) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis")
	return &RedisClient{Client: client}
}

func (r *RedisClient) Close() {
	r.Client.Close()
	log.Println("Redis connection closed")
}
