// SPDX-License-Identifier: GPL-2.0
package redis_client

import (
	"fmt"
	"os"
	"time"
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

func GetRedisObj(redisKey string) (interface{}, error) {
	redisConn := GetRedisConnection()
	defer redisConn.Close()
	jsonStr, err := redisConn.Get(string(redisKey)).Result()
	if err != nil {
		return nil, err
	}
	return jsonStr, nil
}

func SetRedisObj(redisKey string, data interface{}, ttl uint8) error {
	redisConn := GetRedisConnection()
	defer redisConn.Close()
	err := redisConn.Set(string(redisKey), data, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func DeleteRedisObj(redisKeys ...string) error {
	redisConn := GetRedisConnection()
	defer redisConn.Close()
	_, err := redisConn.Del(redisKeys...).Result()
	if err != nil {
		return err
	}
	return nil
}
