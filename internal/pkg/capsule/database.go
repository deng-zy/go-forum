package capsule

import (
	"fmt"
	"forum/internal/pkg/config"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var prefix = "database"
var dbConnections sync.Map

func init() {
	config.Load()
}

func newConnection(name string) (*gorm.DB, error) {
	dsn := getDSN(name)
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, "get sqlDB fail.")
	}
	sqlDB.SetMaxIdleConns(viper.GetInt(fmt.Sprintf("%s.maxIdleConns", prefix)))
	sqlDB.SetMaxOpenConns(viper.GetInt(fmt.Sprintf("%s.maxOpenCoons", prefix)))
	sqlDB.SetConnMaxLifetime(viper.GetDuration(fmt.Sprintf("%s.connMaxLifetime", prefix)) * time.Hour)
	fmt.Printf("+%v", sqlDB.Stats())

	dbConnections.Store(name, db)
	return db, nil
}

func DBConn(args ...string) *gorm.DB {
	name := "default"
	if len(args) > 0 {
		name = args[0]
	}

	connection, ok := dbConnections.Load(name)
	if ok {
		return connection.(*gorm.DB)
	}

	conn, err := newConnection(name)
	if err != nil {
		panic(err)
	}

	dbConnections.Store(name, conn)
	return conn
}

func getDSN(name string) string {
	keyPrefix := fmt.Sprintf("%s.connections.%s", prefix, name)
	username := viper.GetString(fmt.Sprintf("%s.username", keyPrefix))
	password := viper.GetString(fmt.Sprintf("%s.password", keyPrefix))
	host := viper.GetString(fmt.Sprintf("%s.host", keyPrefix))
	port := viper.GetString(fmt.Sprintf("%s.port", keyPrefix))
	charset := viper.GetString(fmt.Sprintf("%s.charset", keyPrefix))
	database := viper.GetString(fmt.Sprintf("%s.database", keyPrefix))

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local", username, password, host, port, database, charset)
}
