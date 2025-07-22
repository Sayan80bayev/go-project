package middleware

import (
	"auth_service/internal/config"
	"auth_service/internal/jwks"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func JWT(cfg *config.Config, log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			log.Warn("Нет токена")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "нет токена"})
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, jwks.JWKS.Keyfunc)
		if err != nil || !token.Valid {
			log.Warnf("Невалидный токен: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "невалидный токен"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("claims", claims)
		c.Next()
	}
}
