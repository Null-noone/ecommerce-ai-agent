package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/spf13/viper"
)

type Config struct {
	rest.RestConf
	Auth         AuthConfig
	Mysql        MysqlConfig
	Redis        RedisConfig
	Kafka        KafkaConfig
	PythonService zrpc.RpcClientConf
}

type AuthConfig struct {
	Secret string
	Expire int
}

type MysqlConfig struct {
	DataSource string
}

type RedisConfig struct {
	Host string
	Port int
	Pass string
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
}

func LoadConfig(path string) *Config {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()
	
	var cfg Config
	viper.Unmarshal(&cfg)
	return &cfg
}
