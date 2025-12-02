package config

import (
	"github.com/Sayan80bayev/go-project/pkg/logging"
	"github.com/spf13/viper"
)

type Config struct {
	Port string `mapstructure:"PORT"`

	RedisAddr string `mapstructure:"REDIS_ADDR"`
	RedisPass string `mapstructure:"REDIS_PASS"`

	RabbitMQUser       string `mapstructure:"RABBIT_MQ_USER"`
	RabbitMQPassword   string `mapstructure:"RABBIT_MQ_PASSWORD"`
	RabbitMQHost       string `mapstructure:"RABBIT_MQ_HOST"`
	RabbitMQPort       string `mapstructure:"RABBIT_MQ_PORT"`
	RabbitMQVHost      string `mapstructure:"RABBIT_MQ_VHOST"`
	RabbitMQQueue      string `mapstructure:"RABBIT_MQ_QUEUE"`
	RabbitMQExchange   string `mapstructure:"RABBIT_MQ_EXCHANGE"`
	RabbitMQRoutingKey string `mapstructure:"RABBIT_MQ_ROUTING_KEY"`

	KeycloakURL   string `mapstructure:"KEYCLOAK_URL"`
	KeycloakRealm string `mapstructure:"KEYCLOAK_REALM"`

	PostgresHost     string `mapstructure:"POSTGRES_HOST"`
	PostgresPort     string `mapstructure:"POSTGRES_PORT"`
	PostgresUser     string `mapstructure:"POSTGRES_USER"`
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDBName   string `mapstructure:"POSTGRES_DB_NAME"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile("config/config.yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logging.Instance.Errorf("Couldn't load config.yaml: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
