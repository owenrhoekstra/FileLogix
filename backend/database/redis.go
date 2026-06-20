package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"FileLogix/utilities/logger"
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

	addr := fmt.Sprintf("%s:%s", host, port)

	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	if err := RDB.Ping(context.Background()).Err(); err != nil {
		logger.Errorf(uuid.Nil, uuid.Nil, "failed to connect to Redis at %s: %v", addr, err)
		log.Fatal(err)
	}

	logger.Infof(uuid.Nil, uuid.Nil, "connected to Redis at %s", addr)
}
