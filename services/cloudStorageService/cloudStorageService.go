package cloudStorageService

import (
	"context"
	"fmt"
	"i9chat/appGlobals"
	"os"
)

func UploadFile(filePath string, data []byte) (string, error) {
	bucketName := os.Getenv("GCS_BUCKET")
	fileUrl := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, filePath)

	stWriter := appGlobals.GCSClient.Bucket(bucketName).Object(filePath).NewWriter(context.Background())

	stWriter.Write(data)

	err := stWriter.Close()
	if err != nil {
		return "", err
	}

	return fileUrl, nil
}
