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
	"utils/appglobals"
)

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

func MapToStruct(mapData map[string]any, structData any) {
	bt, _ := json.Marshal(mapData)

	json.Unmarshal(bt, structData)
}

func AppError(code int, err error) map[string]any {
	if errors.Is(err, appglobals.ErrInternalServerError) {
		return map[string]any{"code": 500, "error": appglobals.ErrInternalServerError.Error()}
	}

	return map[string]any{"code": code, "error": err.Error()}
}

func UploadFile(filePath string, data []byte) (string, error) {
	fileUrl := fmt.Sprintf("https://storage.cloud.google.com/i9chat-bucket/%s", filePath)

	stWriter := appglobals.GCSClient.Bucket("i9chat-bucket").Object(filePath).NewWriter(context.Background())

	stWriter.Write(data)

	err := stWriter.Close()
	if err != nil {
		return "", err
	}

	return fileUrl, nil
}
