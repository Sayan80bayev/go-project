package router

import (
	"engagementService/internal/bootstrap"
	"engagementService/internal/delivery"
	"github.com/Sayan80bayev/go-project/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func SetupSubscriptionRoutes(r *gin.Engine, c *bootstrap.Container) {
	h := delivery.NewSubscriptionHandler(c.SubscriptionService)

	routes := r.Group("api/v1/sub", middleware.AuthMiddleware(c.JWKSUrl))
	{
		routes.POST("/:followeeId/follow", h.Follow)
		routes.DELETE("/:followeeId/unfollow", h.Unfollow)
		routes.GET("/:userId/followers", h.GetFollowers)
		routes.GET("/:userId/following", h.GetFollowing)
	}
}
