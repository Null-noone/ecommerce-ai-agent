package svc

import (
	"ecommerce-ai-agent/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/spf13/viper"
)

type ServiceContext struct {
	Config         *config.Config
	DB             sqlx.DB
	KafkaProducer  KafkaProducer
}

func NewServiceContext() *ServiceContext {
	// Load config from file
	viper.SetConfigFile("etc/ecommerce-api.yaml")
	viper.AutomaticEnv()

	var cfg config.Config
	viper.Unmarshal(&cfg)

	// Database connection
	dsn := cfg.Database.User + ":" + cfg.Database.Password + "@tcp(" + 
		cfg.Database.Host + ":" + string(rune(cfg.Database.Port)) + ")/" + 
		cfg.Database.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
	
	db, _ := sqlx.NewSqlConnection("mysql", dsn)

	return &ServiceContext{
		Config:        &cfg,
		DB:            db,
		KafkaProducer: NewKafkaProducer(cfg.Kafka.Brokers),
	}
}
