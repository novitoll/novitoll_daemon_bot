package redis_client

import (
	"os"
	"fmt"
	"github.com/go-redis/redis"
)

var (
	host = "localhost"
	port = "6379"
)

func RedisClient() *redis.Client {

	if h, ok := os.LookupEnv("REDIS_HOST"); ok {
		host = h
	}
	if p, ok := os.LookupEnv("REDIS_PORT"); ok {
		port = p
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return client
}
