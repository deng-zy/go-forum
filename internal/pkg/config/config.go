package config

import (
	"errors"

	"github.com/spf13/viper"
)

func Load() {
	setDefault()
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("./")
	viper.AddConfigPath("./config")
	viper.SetConfigName("app")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		panic(errors.New(err.Error()))
	}
}

func setDefault() {
	viper.SetDefault("app.host", "127.0.0.1")
	viper.SetDefault("app.port", "8300")
	viper.SetConfigType("yaml")
	viper.SetDefault("app.env", "PROD")
	viper.SetDefault("database.default.host", "127.0.0.1")
	viper.SetDefault("database.default.port", "3306")
	viper.SetDefault("database.default.username", "root")
	viper.SetDefault("database.default.password", "123")
	viper.SetDefault("database.default.database", "task")
	viper.SetDefault("database.maxOpenCoons", 64)
	viper.SetDefault("database.maxIdleConns", 64)
	viper.SetDefault("database.connMaxLifetime", 1)
}
