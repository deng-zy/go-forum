package capsule

import (
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var redisClients sync.Map

//RedisClient new redis client instance
func RedisClient(args ...string) (client *redis.Client) {
	name := "default"
	if len(args) > 0 {
		name = args[0]
	}

	connection, ok := redisClients.Load(name)
	if ok {
		client = connection.(*redis.Client)
		return
	}

	client = newRedisClient(name)
	redisClients.Store(name, client)
	return client
}

//newRedisClient create a redis client
func newRedisClient(name string) *redis.Client {
	prefix := fmt.Sprintf("redis.connections.%s", name)
	options := getRedisOptions(prefix)

	return redis.NewClient(options)
}

//getRedisOptions get redis client option by configure
func getRedisOptions(prefix string) *redis.Options {

	host := viper.GetString(fmt.Sprintf("%s.host", prefix))
	port := viper.GetString(fmt.Sprintf("%s.port", prefix))
	addr := fmt.Sprintf("%s:%s", host, port)

	password := viper.GetString(fmt.Sprintf("%s.password", prefix))
	username := viper.GetString(fmt.Sprintf("%s.username", prefix))

	options := &redis.Options{
		Addr: addr,
		DB:   viper.GetInt(fmt.Sprintf("%s.db", prefix)),
	}

	if password != "" {
		options.Password = password
	}

	if username != "" {
		options.Username = username
	}

	return options
}
