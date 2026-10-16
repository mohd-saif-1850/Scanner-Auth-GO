package rediscache

import (
	"context"

	redisClient "github.com/redis/go-redis/v9"

	"scannerAuth/internal/config"
)

func Connect(cfg *config.Config) (*redisClient.Client, error) {
	options, err := redisClient.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, err
	}

	client := redisClient.NewClient(options)

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
