package cache

import (
	"github.com/redis/go-redis/v9"
)

// Connect creates a new Redis client configured with the given address.
func Connect(redisURL string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
}
