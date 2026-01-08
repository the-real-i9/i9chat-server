package gcsHelpers

import (
	"context"
	"errors"
	"i9chat/src/appGlobals"
	"i9chat/src/helpers"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

func GetMediaurl(mcn string) (string, error) {
	url, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).SignedURL(mcn, &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add((6 * 24) * time.Hour),
	})
	if err != nil {
		helpers.LogError(err)
		return "", err
	}

	return url, nil
}

func GetMediaInfo(ctx context.Context, mcn string) *storage.ObjectAttrs {
	mInfo, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(mcn).Attrs(ctx)
	if err != nil {
		if !errors.Is(err, storage.ErrObjectNotExist) {
			helpers.LogError(err)
		}

		return nil
	}

	return mInfo
}

func DeleteCloudMedia(ctx context.Context, mcn string) {
	err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(mcn).Delete(ctx)
	if err != nil {
		helpers.LogError(err)
	}
}
