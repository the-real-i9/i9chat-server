package helpers

import (
	"encoding/json"
	"i9chat/appTypes"

	"github.com/gofiber/fiber/v2"
)

func MapToStruct(val map[string]any, yourStruct any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, yourStruct)
}

func StructToMap(val any, yourMap *map[string]any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, yourMap)
}

func ToStruct(val any, structData any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, structData)
}

func ErrResp(err error) appTypes.WSResp {

	errCode := fiber.StatusInternalServerError

	if ferr, ok := err.(*fiber.Error); ok {
		errCode = ferr.Code
	}

	return appTypes.WSResp{StatusCode: errCode, Error: err.Error()}
}
