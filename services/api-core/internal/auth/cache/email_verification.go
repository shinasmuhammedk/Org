package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var rdb *redis.Client

// Initialize Redis (call this in main)
func InitRedis(addr string) {
	rdb = redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

// Store token → userID
func StoreVerificationToken(token string, userID string, ttl time.Duration) error {
	key := "email_verify:" + token
	return rdb.Set(ctx, key, userID, ttl).Err()
}

// Get userID from token
func GetUserIDByToken(token string) (string, error) {
	key := "email_verify:" + token
	return rdb.Get(ctx, key).Result()
}

// Delete token after use
func DeleteToken(token string) error {
	key := "email_verify:" + token
	return rdb.Del(ctx, key).Err()
}