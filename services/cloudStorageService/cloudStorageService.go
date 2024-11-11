package cloudStorageService

import (
	"context"
	"fmt"
	"i9chat/appGlobals"
)

func UploadFile(filePath string, data []byte) (string, error) {
	fileUrl := fmt.Sprintf("https://storage.googleapis.com/i9chat-bucket/%s", filePath)

	stWriter := appGlobals.GCSClient.Bucket("i9chat-bucket").Object(filePath).NewWriter(context.Background())

	stWriter.Write(data)

	err := stWriter.Close()
	if err != nil {
		return "", err
	}

	return fileUrl, nil
}
