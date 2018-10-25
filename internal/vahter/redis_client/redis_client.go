// SPDX-License-Identifier: GPL-2.0
package redis_client

import (
	"fmt"
	"os"
	"runtime"

	"github.com/go-redis/redis"
)

const (
	REDIS_MAX_ACTIVE_CONNECTIONS = 1000
)

var (
	host = "localhost"
	port = "6379"
)

func init() {
	if h, ok := os.LookupEnv("REDIS_HOST"); ok {
		host = h
	}
	if p, ok := os.LookupEnv("REDIS_PORT"); ok {
		port = p
	}
}

func GetRedisConnection() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:       fmt.Sprintf("%s:%s", host, port),
		Password:   "", // no password set
		DB:         0,  // use default DB
		MaxRetries: 3,
		PoolSize:   runtime.NumCPU() * 10, // TODO: need to calculate more carefully with ulimit and need to have a Pool of connections
	})
	return client
}
