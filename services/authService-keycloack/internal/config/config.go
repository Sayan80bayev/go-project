package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port             string `mapstructure:"PORT"`
	KeycloakURL      string `mapstructure:"KEYCLOAK_URL"`
	KeycloakRealm    string `mapstructure:"KEYCLOAK_REALM"`
	ClientID         string `mapstructure:"KEYCLOAK_CLIENT_ID"`
	ClientSecret     string `mapstructure:"KEYCLOAK_CLIENT_SECRET"`
	KeycloakCertPath string `mapstructure:"KEYCLOAK_CERT_PATH"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile("config/config.yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("⚠️ Не удалось загрузить config.yaml: %v (продолжим с переменными окружения)", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Фолбэк по умолчанию
	if cfg.Port == "" {
		cfg.Port = "8082"
		log.Println("⚠️ PORT не найден, используется значение по умолчанию: 8082")
	}

	return &cfg, nil
}
