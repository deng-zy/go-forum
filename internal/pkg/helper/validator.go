package helper

import (
	"errors"
	"forum/internal/pkg/constants"
	"forum/pkg/res"
	"net/http"

	"github.com/gin-gonic/gin"
)

var ErrInvalidParams = errors.New("invalid params")

func Validator(c *gin.Context, obj interface{}) bool {
	err := c.ShouldBind(obj)
	if err != nil {
		c.JSON(http.StatusOK, res.JsonErrorMessage(constants.InvalidParams, ErrInvalidParams.Error()))
		c.Abort()
		return false
	}
	return true
}
