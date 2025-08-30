package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Sayan80bayev/go-project/pkg/events"
	"github.com/Sayan80bayev/go-project/pkg/logging"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

type ConsumerConfig struct {
	BootstrapServers string
	GroupID          string
	Topics           []string
}

// Generic event handler function type
type EventHandler func(json.RawMessage) error

type KafkaConsumer struct {
	config   ConsumerConfig
	consumer *kafka.Consumer
	handlers map[string]EventHandler
	log      *logrus.Logger
}

// NewKafkaConsumer creates a new KafkaConsumer instance
func NewKafkaConsumer(cfg ConsumerConfig) (*KafkaConsumer, error) {
	logger := logging.GetLogger()

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":     cfg.BootstrapServers,
		"group.id":              cfg.GroupID,
		"auto.offset.reset":     "earliest",
		"broker.address.family": "v4",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	consumer := &KafkaConsumer{
		config:   cfg,
		consumer: c,
		handlers: make(map[string]EventHandler),
		log:      logger,
	}

	return consumer, nil
}

// RegisterHandler binds a handler to an event type
func (c *KafkaConsumer) RegisterHandler(eventType string, handler EventHandler) {
	c.handlers[eventType] = handler
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	if err := c.consumer.SubscribeTopics(c.config.Topics, nil); err != nil {
		c.log.Errorf("Error subscribing to topics: %v", err)
		return
	}

	c.log.Infof("KafkaConsumer started on topics: %v", c.config.Topics)

	for {
		select {
		case <-ctx.Done():
			c.log.Info("KafkaConsumer stopped by context cancellation")
			return
		default:
			msg, err := c.consumer.ReadMessage(-1)
			if err == nil {
				c.log.Infof("Received message: %s", string(msg.Value))
				c.handleMessage(msg)
			} else {
				c.log.Warnf("KafkaConsumer error: %v", err)
			}
		}
	}
}

func (c *KafkaConsumer) Close() {
	if err := c.consumer.Close(); err != nil {
		c.log.Errorf("Could not close consumer connection gracefully: %v", err)
	} else {
		c.log.Info("Kafka consumer closed gracefully")
	}
}

func (c *KafkaConsumer) handleMessage(msg *kafka.Message) {
	var event events.Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		c.log.Errorf("Error parsing message: %v", err)
		return
	}

	if handler, ok := c.handlers[event.Type]; ok {
		if err := handler(event.Data); err != nil {
			c.log.Errorf("Handler for event %s failed: %v", event.Type, err)
		}
	} else {
		c.log.Warnf("No handler registered for event type: %s", event.Type)
	}
}