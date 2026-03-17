package config

import "os"

type Config struct {
	AppPort string

	// Postgres
	PostgresDSN string

	// Kafka
	KafkaBroker string
	KafkaTopic  string
	KafkaGroup  string

	// Redis
	RedisAddr string
}

func Load() *Config {
	return &Config{
		AppPort:     getEnv("APP_PORT", "8080"),
		PostgresDSN: getEnv("POSTGRES_DSN", "host=postgres user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"),

		KafkaBroker: getEnv("KAFKA_BROKER", "kafka:9092"),
		KafkaTopic:  getEnv("KAFKA_TOPIC", "demo-events"),
		KafkaGroup:  getEnv("KAFKA_GROUP", "monitoring-demo-group"),
		RedisAddr:   getEnv("REDIS_ADDR", "redis:6379"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
