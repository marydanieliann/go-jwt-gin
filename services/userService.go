package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go-jwt/initializers"
	"go-jwt/models"
	"time"
)

func CacheUser(ctx context.Context, user models.User) error {
	userData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to serialize user: %v", err)
	}

	key := "user:" + user.Email
	return initializers.RedisClient.Set(ctx, key, userData, time.Hour*24*30).Err()
}

func GetUserFromCache(ctx context.Context, email string) (*models.User, error) {
	userData, err := initializers.RedisClient.Get(ctx, "user:"+email).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("user not found in cache")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user from cache: %v", err)
	}

	var user models.User
	err = json.Unmarshal([]byte(userData), &user)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize user data: %v", err)
	}
	return &user, nil
}

func InvalidateUserCache(ctx context.Context, email string) error {
	return initializers.RedisClient.Del(ctx, "user:"+email).Err()
}
