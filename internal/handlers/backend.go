package handlers

import (
	"forum/internal/handlers/backend"

	"github.com/gin-gonic/gin"
)

func backendRouter(r *gin.RouterGroup) {
	r.GET("/welcome", backend.Hello)
}
