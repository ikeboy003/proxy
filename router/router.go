package router

import (
	"forwardproxy/router/nba"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures the Gin router
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// NBA routes
	r.POST("/nba", nba.HandleNBA)

	return r
}
