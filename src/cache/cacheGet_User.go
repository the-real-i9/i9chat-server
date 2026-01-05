package cache

import (
	"context"
	"i9chat/src/helpers"

	"github.com/redis/go-redis/v9"
)

// Do here
func GetUser[T any](ctx context.Context, username string) (user T, err error) {
	userJson, err := rdb().HGet(ctx, "users", username).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return user, err
	}

	return helpers.FromJson[T](userJson), nil
}
