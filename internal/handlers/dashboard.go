package handlers

import (
	"forum/internal/handlers/dashboard"

	"github.com/gin-gonic/gin"
)

func dashboardRouter(r *gin.RouterGroup) {
	r.POST("forum", dashboard.CreateForum)
	r.GET("forum", dashboard.AllForum)
	r.DELETE("forum/:forum", dashboard.DeleteForum)
}
