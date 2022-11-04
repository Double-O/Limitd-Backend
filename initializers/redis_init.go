package initializers

import (
	"os"

	"github.com/go-redis/redis/v9"
)

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_DSN"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}
