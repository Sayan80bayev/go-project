package router

import (
	"auth_service/internal/config"
	"auth_service/internal/handler"
	"auth_service/internal/middleware"
	"auth_service/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Setup(cfg *config.Config, log *logrus.Logger) *gin.Engine {

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = log.Out

	r := gin.New()
	r.Use(logging.Middleware)

	api := r.Group("/api/v1")

	// Публичные маршруты
	api.POST("/login", handler.Login(cfg, log))
	api.POST("/refresh", handler.Refresh(cfg, log))
	api.POST("/register", handler.Register(cfg, log))
	api.POST("/logout", handler.Logout(cfg, log))

	// Защищённые маршруты
	auth := api.Group("/auth")
	auth.Use(middleware.JWT(cfg, log))
	auth.GET("/profile", handler.Profile())

	return r
}
