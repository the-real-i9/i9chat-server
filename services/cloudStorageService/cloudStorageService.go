package cloudStorageService

import (
	"context"
	"errors"
	"fmt"
	"i9chat/appGlobals"
	"log"
	"os"
	"time"
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

		return "", appGlobals.ErrInternalServerError
	}

	return fileUrl, nil
}
