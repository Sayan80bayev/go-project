package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/Sayan80bayev/go-project/pkg/logging"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducer struct {
	producer *kafka.Producer
	topic    string
	log      *logrus.Logger
}

// NewKafkaProducer creates a new KafkaProducer instance
func NewKafkaProducer(brokers, topic string) (*KafkaProducer, error) {
	logger := logging.GetLogger()

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		logger.Warnf("Failed to create Kafka producer: %v", err)
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	logger.Infof("Kafka producer initialized for topic: %s", topic)

	return &KafkaProducer{
		producer: p,
		topic:    topic,
		log:      logger,
	}, nil
}

func (p *KafkaProducer) Produce(ctx context.Context, eventType string, data interface{}) error {
	select {
	case <-ctx.Done():
		p.log.Warnf("Produce cancelled by context: %v", ctx.Err())
		return ctx.Err()
	default:
	}

	event := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: eventType,
		Data: data,
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		p.log.Warnf("Failed to marshal event: %v", err)
		return fmt.Errorf("marshal event failed: %w", err)
	}

	err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Value:          jsonData,
	}, nil)
	if err != nil {
		p.log.Warnf("Failed to produce message: %v", err)
		return fmt.Errorf("produce message failed: %w", err)
	}

	p.log.Infof("Message produced to topic %s: %s", p.topic, string(jsonData))
	return nil
}

func (p *KafkaProducer) Close() {
	p.producer.Flush(5000)
	p.producer.Close()
	p.log.Info("Kafka producer closed gracefully")
}
