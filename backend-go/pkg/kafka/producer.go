package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/threading"
)

type Producer interface {
	SendMessage(ctx context.Context, topic, key string, value interface{}) error
	SendOrderCreatedEvent(orderID, userID uint, totalAmount float64) error
	Close() error
}

type SaramaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewProducer(brokers []string, topic string) (Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewHashPartitioner

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &SaramaProducer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *SaramaProducer) SendMessage(ctx context.Context, topic, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}

type OrderCreatedEvent struct {
	Event       string  `json:"event"`
	OrderID     uint    `json:"order_id"`
	UserID      uint    `json:"user_id"`
	TotalAmount float64 `json:"total_amount"`
	Timestamp   int64   `json:"timestamp"`
}

func (p *SaramaProducer) SendOrderCreatedEvent(orderID, userID uint, totalAmount float64) error {
	event := OrderCreatedEvent{
		Event:       "order_created",
		OrderID:     orderID,
		UserID:      userID,
		TotalAmount: totalAmount,
		Timestamp:   time.Now().Unix(),
	}

	// Send asynchronously
	threading.GoSafe(func() {
		p.SendMessage(context.Background(), p.topic, fmt.Sprintf("%d", orderID), event)
	})

	return nil
}

func (p *SaramaProducer) Close() error {
	return p.producer.Close()
}

// Consumer for processing messages
type Consumer struct {
	client  sarama.ConsumerGroup
	handlers map[string]MessageHandler
}

type MessageHandler func(key, value []byte) error

func NewConsumer(brokers []string, groupID string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	client, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		client:   client,
		handlers: make(map[string]MessageHandler),
	}, nil
}

func (c *Consumer) RegisterHandler(topic string, handler MessageHandler) {
	c.handlers[topic] = handler
}

func (c *Consumer) Start(ctx context.Context) error {
	topics := make([]string, 0, len(c.handlers))
	for topic := range c.handlers {
		topics = append(topics, topic)
	}

	consumerHandler := &consumerGroupHandler{handlers: c.handlers}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := c.client.Consume(ctx, topics, consumerHandler)
			if err != nil {
				return err
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.client.Close()
}

type consumerGroupHandler struct {
	handlers map[string]MessageHandler
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if handler, ok := h.handlers[msg.Topic]; ok {
			handler(msg.Key, msg.Value)
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
