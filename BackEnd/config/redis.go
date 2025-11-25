package config

import (
	"context"
	"exchangeapp/global"
	"log"

	"github.com/redis/go-redis/v9"
)

func InitRedis() {
	ctx := context.Background()
	RedisClient := redis.NewClient(&redis.Options{
		Addr:     AppConfig.Redis.Addr,
		Password: AppConfig.Redis.Password,
		DB:       AppConfig.Redis.DB,
	})

	_, err := RedisClient.Ping(ctx).Result()

	if err != nil {
		log.Fatalf("Failed to connect to Redis,got error: %v", err)
	}

	global.RedisDB = RedisClient
}
