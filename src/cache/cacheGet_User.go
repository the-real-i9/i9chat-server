package cache

import (
	"context"
	"i9chat/src/helpers"
	"i9chat/src/services/cloudStorageService"

	"github.com/redis/go-redis/v9"
)

func GetUser[T any](ctx context.Context, username string) (user T, err error) {
	userJson, err := rdb().HGet(ctx, "users", username).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return user, err
	}

	userMap := helpers.FromJson[map[string]any](userJson)

	if err := cloudStorageService.ProfilePicCloudNameToUrl(userMap); err != nil {
		return user, err
	}

	return helpers.ToStruct[T](userMap), nil
}
