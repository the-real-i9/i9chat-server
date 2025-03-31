package cloudStorageService

import (
	"context"
	"errors"
	"fmt"
	"i9chat/src/appGlobals"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Upload(ctx context.Context, filePath string, data []byte) (string, error) {
	mediaUploadCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	bucketName := os.Getenv("GCS_BUCKET")
	fileUrl := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, filePath)

	stWriter := appGlobals.GCSClient.Bucket(bucketName).Object(filePath).NewWriter(mediaUploadCtx)

	stWriter.Write(data)

	err := stWriter.Close()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println("cloudStorageService.go: UploadFile:", "media upload timed out")
		} else {
			log.Println("cloudStorageService.go: UploadFile:", err)
		}

		return "", fiber.ErrInternalServerError
	}

	return fileUrl, nil
}
