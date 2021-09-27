package app

import (
	"fmt"
	"forum/internal/handlers"
	"forum/internal/pkg/config"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func init() {
	config.Load()
}

func Run() {
	app := gin.New()
	addr := fmt.Sprintf("%s:%s", viper.GetString("app.host"), viper.GetString("app.port"))

	// app.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	app.Use(gin.Logger())
	pprof.Register(app, "optimal")
	handlers.Router(app)

	app.Run(addr)
}
