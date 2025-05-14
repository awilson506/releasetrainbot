package routes

import (
	"github.com/awilson506/releasetrain/config"
	_ "github.com/awilson506/releasetrain/docs"
	"github.com/awilson506/releasetrain/handlers"
	"github.com/awilson506/releasetrain/slackbolt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes initializes the routes for the application
func SetupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	slackHandler := slackbolt.NewSlackCommandHandler()

	// Registering the /release-train command
	slackHandler.RegisterCommand("/release-train", handlers.SlackReleaseTrainHandler)

	if config.IsCloudfrontEnabled() {
		// Middleware to check CloudFront token
		router.Use(checkCloudFrontToken())
	}

	// POST request for Slack commands
	v1 := router.Group("/v1")
	{
		v1.Use(slackMiddleware())
		// Slack command handler
		v1.POST("/slack/command", slackHandler.Handle)
	}

	// Swagger docs
	if !config.IsProduction() {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}
