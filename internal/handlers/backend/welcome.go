package backend

import (
	"forum/pkg/res"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Hello(c *gin.Context) {
	c.JSON(http.StatusOK, res.JsonSuccess())
}
