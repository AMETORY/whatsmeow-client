package service

import (
	"github.com/go-redis/redis"
)

var REDIS = &redis.Client{}

func InitRedis() {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	REDIS = client
}
