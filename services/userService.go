package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go-jwt/initializers"
	"go-jwt/models"
)

func UserToCache(ctx context.Context, user models.User) error {
	userData, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("error marshaling user data: %w", err)
	}
	err = initializers.RedisClient.HSet(ctx, "Users", user.Email, userData).Err()
	if err != nil {
		return fmt.Errorf("error saving user to cache: %w", err)
	}

	return nil
}

func GetUserFromCache(ctx context.Context, email string) (*models.User, error) {
	userData, err := initializers.RedisClient.HGet(ctx, "Users", email).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("user not found in cache for email: %s", email)
		}
		return nil, fmt.Errorf("error retrieving user from cache: %w", err)
	}

	fmt.Println("Raw user data from Redis:", userData)

	var user models.User
	err = json.Unmarshal([]byte(userData), &user)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling user data: %w", err)
	}
	return &user, nil
}
