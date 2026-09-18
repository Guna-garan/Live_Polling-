package database

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis dials Redis using a standard redis:// URL and verifies the
// connection with a PING before returning.
func ConnectRedis(url string) (*redis.Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("redis: connected")
	return client, nil
}
