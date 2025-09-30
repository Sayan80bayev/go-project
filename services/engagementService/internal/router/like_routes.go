package router

import (
	"engagementService/internal/bootstrap"
	"engagementService/internal/delivery"
	"github.com/Sayan80bayev/go-project/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func SetupLikeRoutes(r *gin.Engine, c *bootstrap.Container) {
	h := delivery.NewLikeHandler(c.LikeService)
	routes := r.Group("api/v1/like", middleware.AuthMiddleware(c.JWKSUrl))
	{
		routes.POST("/:postId/like", h.Like)
		routes.DELETE("/:postId/unlike", h.Unlike)
		routes.GET("/user/:userId/likes", h.GetUserLikes)
		routes.GET("/post/:postId/likes", h.GetPostLikes)
	}
}
