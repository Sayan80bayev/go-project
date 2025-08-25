package middleware

import (
	"github.com/google/uuid"
	"log"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwksURL string) gin.HandlerFunc {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshUnknownKID: true,
	})
	if err != nil {
		log.Fatalf("Could not load JWKS: %v", err)
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

		if subStr, ok := claims["sub"].(string); ok {
			if subUUID, err := uuid.Parse(subStr); err == nil {
				c.Set("user_id", subUUID)
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in token"})
				return
			}
		}

		if username, ok := claims["preferred_username"].(string); ok {
			c.Set("username", username)
		}

		// Map raw roles from Keycloak → our app roles
		if resourceAccess, ok := claims["resource_access"].(map[string]interface{}); ok {
			if authService, ok := resourceAccess["auth_service"].(map[string]interface{}); ok {
				if roles, ok := authService["roles"].([]interface{}); ok {
					var appRoles []Role
					for _, role := range roles {
						if r, ok := role.(string); ok {
							switch strings.ToLower(r) {
							case "admin":
								appRoles = append(appRoles, RoleAdmin)
							case "moder":
								appRoles = append(appRoles, RoleModerator)
							case "user":
								appRoles = append(appRoles, RoleUser)
							default:
								// ignore unknown roles
							}
						}
					}
					c.Set("roles", appRoles)
				}
			}
		}

		c.Next()
	}
}
