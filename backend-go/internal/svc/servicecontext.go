package svc

import (
	"fmt"

	"ecommerce-ai-agent/internal/config"
	"ecommerce-ai-agent/pkg/kafka"
	"ecommerce-ai-agent/pkg/redis"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config        *config.Config
	DB            *gorm.DB
	Redis         *redis.Client
	KafkaProducer kafka.Producer
	Lock          *redis.DistributedLock
}

func NewServiceContext() *ServiceContext {
	cfg := &config.Config{
		RestConf: config.RestConf{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Auth: config.AuthConfig{
			Secret: "your-secret-key-change-in-production",
			Expire: 86400,
		},
		Mysql: config.MysqlConfig{
			DataSource: "ecom_user:EcomPass456!@tcp(mysql:3306)/ecommerce_db?charset=utf8mb4&parseTime=True&loc=Local",
		},
		Redis: config.RedisConfig{
			Host: "redis",
			Port: 6379,
			Pass: "RedisPass789!",
		},
		Kafka: config.KafkaConfig{
			Brokers: []string{"kafka:9092"},
			Topic:   "order_events",
		},
	}

	// Initialize MySQL
	db, err := gorm.Open(mysql.Open(cfg.Mysql.DataSource), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect to MySQL: %v\n", err)
	}

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Pass,
		DB:       0,
	})

	// Initialize Kafka Producer
	kafkaProducer, _ := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)

	// Initialize Distributed Lock
	lock := redis.NewDistributedLock(redisClient, "order", 10*1000*1000*1000) // 10 seconds

	return &ServiceContext{
		Config:        cfg,
		DB:            db,
		Redis:         redisClient,
		KafkaProducer: kafkaProducer,
		Lock:          lock,
	}
}
