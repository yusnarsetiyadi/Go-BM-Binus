package database

import (
	"bm_binus/internal/config"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

func newRedisClient(host string, password string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       0,
	})
	return client
}

func InitRedis() *redis.Client {
	var redisHost = fmt.Sprintf("%s:%s", config.Get().Redis.RedisHost, config.Get().Redis.RedisPort)
	var redisPassword = config.Get().Redis.RedisPassword

	rdb := newRedisClient(redisHost, redisPassword)
	logrus.Info("REDIS client initialized")

	return rdb
}
