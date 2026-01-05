package cache

import (
	"i9chat/src/appGlobals"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/redis/go-redis/v9"
)

func getMediaurl(mcn string) (string, error) {
	url, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).SignedURL(mcn, &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add((6 * 24) * time.Hour),
	})
	if err != nil {
		return "", err
	}

	return url, nil
}

func rdb() *redis.Client {
	return appGlobals.RedisClient
}
