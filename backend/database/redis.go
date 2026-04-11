package database

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis() {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "6379"
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})

	if err := RDB.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Redis connection failed:", err)
	}

	log.Println("Redis connected at", host+":"+port)
}
