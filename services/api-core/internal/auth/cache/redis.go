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

func StoreRefreshToken(ctx context.Context, token string, userID string) error {
	key := "refresh:" + token
	return rdb.Set(ctx, key, userID, 7*24*time.Hour).Err()
}

func GetRefreshToken(ctx context.Context, token string) (string, error) {
	key := "refresh:" + token
	return rdb.Get(ctx, key).Result()
}

func DeleteRefreshToken(ctx context.Context, token string) error {
	key := "refresh:" + token
	return rdb.Del(ctx, key).Err()
}