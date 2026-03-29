package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppPort string
	AppEnv  string

	// Postgres
	PostgresURL         string
	PostgresMaxConns    int
	PostgresMinConns    int
	PostgresConnTimeout time.Duration

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Kafka
	KafkaBroker      string
	KafkaTopic       string
	KafkaGroupID     string
	KafkaMinBytes    int
	KafkaMaxBytes    int
	KafkaMaxAttempts int
}

func Load() *Config {
	cfg := &Config{
		// App
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),

		// Postgres
		PostgresURL:         getEnv("POSTGRES_URL", "postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable"),
		PostgresMaxConns:    getEnvInt("POSTGRES_MAX_CONNS", 10),
		PostgresMinConns:    getEnvInt("POSTGRES_MIN_CONNS", 2),
		PostgresConnTimeout: getEnvDuration("POSTGRES_CONN_TIMEOUT", 5*time.Second),

		KafkaBroker:      getEnv("KAFKA_BROKER", "kafka:9092"),
		KafkaTopic:       getEnv("KAFKA_TOPIC", "monitoring.events"),
		KafkaGroupID:     getEnv("KAFKA_GROUP_ID", "monitoring-group"),
		KafkaMinBytes:    getEnvInt("KAFKA_MIN_BYTES", 1),       // 1B  - fetch immediately
		KafkaMaxBytes:    getEnvInt("KAFKA_MAX_BYTES", 1048576), // 1MB - max per fetch
		KafkaMaxAttempts: getEnvInt("KAFKA_MAX_ATTEMPTS", 3),    // retry 3 times on failure

		RedisAddr:     getEnv("REDIS_ADDR", "redis:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),
	}
	cfg.log()
	return cfg
}

// log prints config on startup (masks sensitive values)
func (c *Config) log() {
	log.Printf("Config loaded | env=%s port=%s kafka=%s topic=%s postgres=%s redis=%s",
		c.AppEnv,
		c.AppPort,
		c.KafkaBroker,
		c.KafkaTopic,
		maskURL(c.PostgresURL),
		c.RedisAddr,
	)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		log.Printf("Invalid integer for %s: %s, using default: %d", key, val, defaultValue)
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
		log.Printf("Invalid duration for %s: %s, using default: %v", key, val, defaultValue)
	}
	return defaultValue
}

// maskURL hides password in connection strings for safe logging
// postgres://user:PASSWORD@host → postgres://user:***@host
func maskURL(url string) string {
	for i, c := range url {
		if c == ':' {
			for j := i + 1; j < len(url); j++ {
				if url[j] == '@' {
					return url[:i+1] + "***" + url[j:]
				}
			}
		}
	}
	return url
}
