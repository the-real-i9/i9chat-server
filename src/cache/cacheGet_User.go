package cache

import (
	"context"
	"fmt"
	"i9chat/src/helpers"

	"github.com/redis/go-redis/v9"
)

func GetUser[T any](ctx context.Context, username string) (user T, err error) {
	userJson, err := rdb().HGet(ctx, "users", username).Result()
	if err != nil && err != redis.Nil {
		helpers.LogError(err)
		return user, err
	}

	userMap := helpers.FromJson[map[string]any](userJson)

	ppicCloudName := userMap["profile_pic_cloud_name"].(string)

	var (
		smallPPicn  string
		mediumPPicn string
		largePPicn  string
	)

	_, err = fmt.Sscanf(ppicCloudName, "small:%s medium:%s large:%s", &smallPPicn, &mediumPPicn, &largePPicn)
	if err != nil {
		return user, err
	}

	smallPicUrl, err := getMediaurl(smallPPicn)
	if err != nil {
		return user, err
	}

	mediumPicUrl, err := getMediaurl(mediumPPicn)
	if err != nil {
		return user, err
	}

	largePicUrl, err := getMediaurl(largePPicn)
	if err != nil {
		return user, err
	}

	userMap["profile_pic_url"] = fmt.Sprintf("small:%s medium:%s large:%s", smallPicUrl, mediumPicUrl, largePicUrl)

	delete(userMap, "profile_pic_cloud_name")

	return helpers.ToStruct[T](userMap), nil
}
