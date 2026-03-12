package svc

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
)

type KafkaProducer interface {
	SendMessage(topic, key string, value []byte) error
	Close() error
}

type SaramaProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(brokers []string) KafkaProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		fmt.Printf("Failed to create Kafka producer: %v\n", err)
		return nil
	}

	return &SaramaProducer{producer: producer}
}

func (p *SaramaProducer) SendMessage(topic, key string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	_, _, err := p.producer.SendMessage(msg)
	return err
}

func (p *SaramaProducer) Close() error {
	return p.producer.Close()
}
