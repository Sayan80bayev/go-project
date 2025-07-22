package main

import (
	"auth_service/internal/config"
	"auth_service/internal/router"
	"auth_service/pkg/logging"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	log := logging.GetLogger()

	r := router.Setup(cfg, log)

	log.Infof("🚀 Сервер запущен на :%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
