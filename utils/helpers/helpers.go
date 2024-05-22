package helpers

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var ErrInternalServerError = errors.New("internal server error: check logger")

func LoadEnv() error {
	dotenv, err := os.Open(".env")
	if err != nil {
		return err
	}

	env := bufio.NewScanner(dotenv)

	for env.Scan() {
		key, value, found := strings.Cut(env.Text(), "=")
		if !found || strings.HasPrefix(key, "#") {
			continue
		}

		err := os.Setenv(key, value)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

var GCSClient *storage.Client

func InitGCSClient() error {
	stClient, err := storage.NewClient(context.Background(), option.WithCredentialsFile("i9apps-storage.json"))
	if err != nil {
		return err
	}

	GCSClient = stClient

	return nil
}

func ParseToStruct(val any, structData any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, structData)
}

func ErrResp(code int, err error) map[string]any {
	if errors.Is(err, ErrInternalServerError) {
		return map[string]any{"statusCode": 500, "error": ErrInternalServerError.Error()}
	}

	return map[string]any{"statusCode": code, "error": err.Error()}
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
