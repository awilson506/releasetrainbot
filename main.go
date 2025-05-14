package main

import (
	routes "github.com/awilson506/releasetrain/router"
	"github.com/awilson506/releasetrain/scheduler"
	"github.com/gin-gonic/gin"

	"github.com/awilson506/releasetrain/config"
	"github.com/awilson506/releasetrain/db"
)

// @title			Slack Release Train Bot API
// @version		1.0
// @description	This is the API documentation for the Slack rotation bot that manages user rotations.
// @host			localhost:8080
// @BasePath		/v1
func main() {
	config.Load()
	db.Init()

	r := gin.Default()
	r.SetTrustedProxies(nil)
	routes.SetupRoutes(r)

	scheduler.Start()

	// Should update this to use a cert for local dev as well
	if config.IsProduction() {
		r.RunTLS(":443", config.CertFile, config.KeyFile)
	} else {
		r.Run(":8080")
	}
}
