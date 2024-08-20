package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"i9chat/globals"
	"i9chat/utils/appTypes"
)

func MapToStruct(val map[string]any, structData any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, structData)
}

func ToStruct(val any, structData any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, structData)
}

func ErrResp(code int, err error) appTypes.WSResp {
	if errors.Is(err, globals.ErrInternalServerError) {
		return appTypes.WSResp{StatusCode: 500, Error: globals.ErrInternalServerError.Error()}
	}

	return appTypes.WSResp{StatusCode: code, Error: err.Error()}
}

func UploadFile(filePath string, data []byte) (string, error) {
	fileUrl := fmt.Sprintf("https://storage.cloud.google.com/i9chat-bucket/%s", filePath)

	stWriter := globals.GCSClient.Bucket("i9chat-bucket").Object(filePath).NewWriter(context.Background())

	stWriter.Write(data)

	err := stWriter.Close()
	if err != nil {
		return "", err
	}

	return fileUrl, nil
}
