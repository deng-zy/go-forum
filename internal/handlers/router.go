package handlers

import (
	"forum/internal/pkg/config"

	"github.com/gin-gonic/gin"
)

func init() {
	config.Load()
}

func Router(app *gin.Engine) {
	backend := app.Group("/backend/api")
	dashboard := app.Group("/dashboard/api")

	dashboardRouter(dashboard)
	backendRouter(backend)
}
