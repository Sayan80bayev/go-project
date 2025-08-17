package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware проверяет JWT и сохраняет нужные поля в Gin Context
func AuthMiddleware(jwksURL string) gin.HandlerFunc {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshUnknownKID: true,
	})
	if err != nil {
		log.Fatalf("Не удалось загрузить JWKS: %v", err)
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, jwks.Keyfunc)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// Сохраняем нужные поля в контекст
		if sub, ok := claims["sub"].(string); ok {
			c.Set("user_id", sub)
		}
		if username, ok := claims["preferred_username"].(string); ok {
			c.Set("username", username)
		}
		// Из Keycloak роли обычно в claims["realm_access"].roles
		if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
			if roles, ok := realmAccess["roles"].([]interface{}); ok {
				var strRoles []string
				for _, role := range roles {
					if r, ok := role.(string); ok {
						strRoles = append(strRoles, r)
					}
				}
				c.Set("roles", strRoles)
			}
		}

		c.Next()
	}
}
