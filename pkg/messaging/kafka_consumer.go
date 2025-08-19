package messaging

import (
	"encoding/json"
	"sync"

	"github.com/Sayan80bayev/go-project/pkg/events"
	"github.com/Sayan80bayev/go-project/pkg/logging"
	"github.com/confluentinc/confluent-kafka-go/kafka"
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
}

var (
	logger           = logging.GetLogger()
	consumerInstance *KafkaConsumer
	consumerOnce     sync.Once
)

// NewKafkaConsumer creates a new KafkaConsumer instance
func NewKafkaConsumer(cfg ConsumerConfig) (*KafkaConsumer, error) {
	var err error

	var c *kafka.Consumer
	c, err = kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.BootstrapServers,
		"group.id":          cfg.GroupID,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		return nil, err
	}

	consumer := &KafkaConsumer{
		config:   cfg,
		consumer: c,
		handlers: make(map[string]EventHandler),
	}

	return consumer, err
}

// GetKafkaConsumer returns a singleton KafkaConsumer instance
func GetKafkaConsumer(config ConsumerConfig) (*KafkaConsumer, error) {
	var err error
	consumerOnce.Do(func() {
		var c *kafka.Consumer
		c, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers": config.BootstrapServers,
			"group.id":          config.GroupID,
			"auto.offset.reset": "earliest",
		})
		if err != nil {
			return
		}
		consumerInstance = &KafkaConsumer{
			config:   config,
			consumer: c,
			handlers: make(map[string]EventHandler),
		}
	})
	return consumerInstance, err
}

// RegisterHandler binds a handler to an event type
func (c *KafkaConsumer) RegisterHandler(eventType string, handler EventHandler) {
	c.handlers[eventType] = handler
}

func (c *KafkaConsumer) Start() {
	if err := c.consumer.SubscribeTopics(c.config.Topics, nil); err != nil {
		logger.Errorf("Error subscribing to topics: %v", err)
		return
	}

	logger.Info("KafkaConsumer started...")

	for {
		msg, err := c.consumer.ReadMessage(-1)
		if err == nil {
			logger.Infof("Received message: %s", string(msg.Value))
			c.handleMessage(msg)
		} else {
			logger.Warnf("KafkaConsumer error: %v", err)
		}
	}
}

func (c *KafkaConsumer) Close() {
	if err := c.consumer.Close(); err != nil {
		logger.Errorf("Could not close consumer's connection gracefully: %v", err)
	}
}

func (c *KafkaConsumer) handleMessage(msg *kafka.Message) {
	var event events.Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		logger.Errorf("Error parsing message: %v", err)
		return
	}

	if handler, ok := c.handlers[event.Type]; ok {
		if err := handler(event.Data); err != nil {
			logger.Errorf("Handler for event %s failed: %v", event.Type, err)
		}
	} else {
		logger.Warnf("No handler registered for event type: %s", event.Type)
	}
}
