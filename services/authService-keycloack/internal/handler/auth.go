package handler

import (
	"auth_service/internal/config"
	"auth_service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func Login(cfg *config.Config, log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := service.LoginUser(c, cfg)
		if err != nil {
			log.Warnf("Ошибка логина: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

func Refresh(cfg *config.Config, log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := service.RefreshToken(c, cfg)
		if err != nil {
			log.Warnf("Ошибка обновления токена: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

func Register(cfg *config.Config, log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := service.RegisterUser(c, cfg)
		if err != nil {
			log.Errorf("Ошибка регистрации: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "user создан с ролью user"})
	}
}

func Profile() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := c.Get("claims")
		c.JSON(http.StatusOK, gin.H{"claims": claims})
	}
}

func Logout(cfg *config.Config, log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := service.LogoutUser(c, cfg)
		if err != nil {
			log.Warnf("Ошибка logout: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
