package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"i9chat/utils/appTypes"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func initGCSClient() error {
	stClient, err := storage.NewClient(context.Background(), option.WithCredentialsFile("i9apps-storage.json"))
	if err != nil {
		return err
	}

	GCSClient = stClient

	return nil
}

func MapToStruct(val map[string]any, structData any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, structData)
}

func ToStruct(val any, structData any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, structData)
}

func ErrResp(code int, err error) appTypes.WSResp {
	if errors.Is(err, ErrInternalServerError) {
		return appTypes.WSResp{StatusCode: 500, Error: ErrInternalServerError.Error()}
	}

	return appTypes.WSResp{StatusCode: code, Error: err.Error()}
}

func UploadFile(filePath string, data []byte) (string, error) {
	fileUrl := fmt.Sprintf("https://storage.cloud.google.com/i9chat-bucket/%s", filePath)

	stWriter := GCSClient.Bucket("i9chat-bucket").Object(filePath).NewWriter(context.Background())

	stWriter.Write(data)

	err := stWriter.Close()
	if err != nil {
		return "", err
	}

	return fileUrl, nil
}

func InitApp() error {

	godotenv.Load(".env")

	if err := initDBPool(); err != nil {
		return err
	}

	if err := initGCSClient(); err != nil {
		return err
	}

	return nil
}
