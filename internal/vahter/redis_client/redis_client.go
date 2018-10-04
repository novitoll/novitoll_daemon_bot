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

type RedisClient struct {
	Conn *redis.Client
}

func (rc *RedisClient) Connect() {

	if h, ok := os.LookupEnv("REDIS_HOST"); ok {
		host = h
	}
	if p, ok := os.LookupEnv("REDIS_PORT"); ok {
		port = p
	}

	rc.Conn = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}
