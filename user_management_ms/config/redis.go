package config

import "github.com/redis/go-redis/v9"

func ConnectToRedis(url string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{Addr: url})
	return rdb
}
