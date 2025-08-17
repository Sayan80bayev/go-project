package messaging

import (
	"encoding/json"
	"sync"

	"github.com/Sayan80bayev/go-project/pkg/logging"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducer struct {
	producer *kafka.Producer
	topic    string
}

var (
	producerInstance *KafkaProducer
	producerOnce     sync.Once
)

// GetProducer returns a singleton KafkaProducer instance
func GetProducer(brokers, topic string) (*KafkaProducer, error) {
	var err error
	producerOnce.Do(func() {
		var p *kafka.Producer
		p, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers": brokers,
		})
		if err != nil {
			logging.GetLogger().Warnf("Failed to create Kafka producer: %v", err)
			return
		}
		logging.GetLogger().Infof("Kafka producer initialized for topic: %s", topic)
		producerInstance = &KafkaProducer{
			producer: p,
			topic:    topic,
		}
	})
	return producerInstance, err
}

func (p *KafkaProducer) Produce(eventType string, data interface{}) error {
	event := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: eventType,
		Data: data,
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		logging.GetLogger().Warnf("Failed to marshal event: %v", err)
		return err
	}

	err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Value:          jsonData,
	}, nil)
	if err != nil {
		logging.GetLogger().Warnf("Failed to produce message: %v", err)
		return err
	}

	logging.GetLogger().Infof("Message produced to topic %s: %s", p.topic, jsonData)
	return nil
}

func (p *KafkaProducer) Close() {
	p.producer.Flush(5000)
	p.producer.Close()
	logging.GetLogger().Info("Kafka producer closed gracefully")
}
