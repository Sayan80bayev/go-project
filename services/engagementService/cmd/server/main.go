package main

import (
	"engagementService/internal/bootstrap"
	"engagementService/internal/router"
	"github.com/Sayan80bayev/go-project/pkg/logging"
	"github.com/gin-gonic/gin"
)

func main() {
	ctn, err := bootstrap.Init()
	if err != nil {
		panic(err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.Middleware)
	SetupRoutes(r, ctn)

	err = r.Run(":" + ctn.Config.Port)
	if err != nil {
		panic(err)
		return
	}
}

func SetupRoutes(r *gin.Engine, ctn *bootstrap.Container) {
	router.SetupSubscriptionRoutes(r, ctn)
	router.SetupLikeRoutes(r, ctn)
}
