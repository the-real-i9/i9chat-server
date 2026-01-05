package chatUploadService

import (
	"context"
	"fmt"
	"i9chat/src/appGlobals"
	"i9chat/src/helpers"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthDataT struct {
	UploadUrl      string `json:"uploadUrl"`
	MediaCloudName string `json:"mediaCloudName"`
}

func Authorize(ctx context.Context, msgType, mediaMIME string, mediaSize int64) (AuthDataT, error) {
	var res AuthDataT

	mediaCloudName := fmt.Sprintf("uploads/chat/%s/%d%d/%s", msgType, time.Now().Year(), time.Now().Month(), uuid.NewString())

	url, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).SignedURL(
		mediaCloudName,
		&storage.SignedURLOptions{
			Scheme:      storage.SigningSchemeV4,
			Method:      "PUT",
			ContentType: mediaMIME,
			Expires:     time.Now().Add(15 * time.Minute),
			Headers:     []string{fmt.Sprintf("x-goog-content-length-range: 0,%d", mediaSize)},
		},
	)
	if err != nil {
		helpers.LogError(err)
		return AuthDataT{}, fiber.ErrInternalServerError
	}

	res.UploadUrl = url
	res.MediaCloudName = mediaCloudName

	return res, nil
}

func AuthorizeVisual(ctx context.Context, msgType string, mediaMIME [2]string, mediaSize [2]int64) (AuthDataT, error) {
	var res AuthDataT

	for blurPlch0_actual1, size := range mediaSize {

		which := [2]string{"blur_placeholder", "actual"}

		mediaCloudName := fmt.Sprintf("uploads/chat/%s/%d%d/%s-%s", msgType, time.Now().Year(), time.Now().Month(), uuid.NewString(), which[blurPlch0_actual1])

		url, err := appGlobals.GCSClient.Bucket(os.Getenv("GCS_BUCKET_NAME")).SignedURL(
			mediaCloudName,
			&storage.SignedURLOptions{
				Scheme:      storage.SigningSchemeV4,
				Method:      "PUT",
				ContentType: mediaMIME[blurPlch0_actual1],
				Expires:     time.Now().Add(15 * time.Minute),
				Headers:     []string{fmt.Sprintf("x-goog-content-length-range: 0,%d", size)},
			},
		)
		if err != nil {
			helpers.LogError(err)
			return AuthDataT{}, fiber.ErrInternalServerError
		}

		if blurPlch0_actual1 == 0 {
			res.UploadUrl += "blur_placeholder:"
			res.MediaCloudName += "blur_placeholder:"
		} else {
			res.UploadUrl += "actual:"
			res.MediaCloudName += "actual:"
		}

		res.UploadUrl += url
		res.MediaCloudName += mediaCloudName

		if blurPlch0_actual1 == 0 {
			res.UploadUrl += " "
			res.MediaCloudName += " "
		}
	}

	return res, nil
}
