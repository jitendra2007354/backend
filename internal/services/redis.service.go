package services

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	redisCtx    = context.Background()
)

func InitRedis() {
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		redisUrl = "redis://localhost:6379/0"
	}

	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		fmt.Printf("⚠️ Invalid REDIS_URL: %v. Using default.\n", err)
		opt = &redis.Options{Addr: "localhost:6379"}
	}

	RedisClient = redis.NewClient(opt)

	if err := RedisClient.Ping(redisCtx).Err(); err != nil {
		fmt.Printf("⚠️ Redis connection failed: %v. Running in single-instance mode.\n", err)
		RedisClient = nil
	} else {
		fmt.Println("✅ Connected to Redis")
	}
}
