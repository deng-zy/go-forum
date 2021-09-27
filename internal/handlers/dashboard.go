package handlers

import (
	"forum/internal/handlers/dashboard"

	"github.com/gin-gonic/gin"
)

func dashboardRouter(r *gin.RouterGroup) {
	r.POST("forum", dashboard.CreateForum)
	r.GET("forum", dashboard.AllForum)
	r.GET("forum/:forum", dashboard.ShowForum)
	r.DELETE("forum/:forum", dashboard.DeleteForum)
	r.PUT("forum/:forum", dashboard.UpdateForum)
}
