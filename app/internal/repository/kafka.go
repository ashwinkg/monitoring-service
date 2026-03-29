package repository

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"

	"github.com/ashwinkg/monitoring-service/internal/config"
)

type KafkaClient struct {
	Writer *kafka.Writer
	Reader *kafka.Reader
}

func NewKafka(cfg *config.Config) *KafkaClient {
	// Producer (Writer)
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaBroker),
		Topic:    cfg.KafkaTopic,
		Balancer: &kafka.LeastBytes{},
	}

	// Consumer (Reader)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KafkaBroker},
		Topic:   cfg.KafkaTopic,
		GroupID: cfg.KafkaGroupID,
	})

	log.Println("✅ Kafka client initialized")
	return &KafkaClient{Writer: writer, Reader: reader}
}

// Produce sends a message to Kafka
func (k *KafkaClient) Produce(ctx context.Context, key, value string) error {
	return k.Writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: []byte(value),
	})
}

// Consume reads messages from Kafka
func (k *KafkaClient) Consume(ctx context.Context, handler func(key, value string)) {
	go func() {
		for {
			msg, err := k.Reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Kafka read error: %v", err)
				return
			}
			handler(string(msg.Key), string(msg.Value))
		}
	}()
}

func (k *KafkaClient) Close() {
	k.Writer.Close()
	k.Reader.Close()
	log.Println("Kafka connections closed")
}
