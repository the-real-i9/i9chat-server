package cloudStorageService

import (
	"context"
	"errors"
	"fmt"
	"i9chat/src/appErrors/userErrors"
	"i9chat/src/appGlobals"
	"i9chat/src/helpers"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Upload(ctx context.Context, filePath string, data []byte) (string, error) {
	mediaUploadCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	bucketName := os.Getenv("GCS_BUCKET")
	mediaUrl := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, filePath)

	if os.Getenv("GO_ENV") != "production" {
		return mediaUrl, nil
	}

	stWriter := appGlobals.GCSClient.Bucket(bucketName).Object(filePath).NewWriter(mediaUploadCtx)

	stWriter.Write(data)

	err := stWriter.Close()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return "", fiber.NewError(fiber.StatusRequestTimeout, userErrors.MediaUploadTimedOut)
		}

		helpers.LogError(err)
		return "", fiber.ErrInternalServerError
	}

	return mediaUrl, nil
}
