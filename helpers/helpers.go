package helpers

import (
	"encoding/json"
	"i9chat/appTypes"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func MapToStruct(val map[string]any, yourStruct any) {
	bt, _ := json.Marshal(val)

	if err := json.Unmarshal(bt, yourStruct); err != nil {
		log.Println("helpers.go: MapToStruct:", err)
	}
}

func AnyToStruct(val any, yourStruct any) {
	bt, _ := json.Marshal(val)

	if err := json.Unmarshal(bt, yourStruct); err != nil {
		log.Println("helpers.go: AnyToStruct:", err)
	}
}

func StructToMap(val any, yourMap *map[string]any) {
	bt, _ := json.Marshal(val)

	if err := json.Unmarshal(bt, yourMap); err != nil {
		log.Println("helpers.go: StructToMap:", err)
	}
}

func ParseIntLimitOffset(limit, offset string) (int, int, error) {
	limitInt, err := strconv.ParseInt(limit, 10, 0)
	if err != nil {
		return 0, 0, err
	}

	offsetInt, err := strconv.ParseInt(offset, 10, 0)
	if err != nil {
		return 0, 0, err
	}

	return int(limitInt), int(offsetInt), nil
}

func ErrResp(err error) appTypes.WSResp {

	errCode := fiber.StatusInternalServerError

	if ferr, ok := err.(*fiber.Error); ok {
		errCode = ferr.Code
	}

	return appTypes.WSResp{StatusCode: errCode, Error: err.Error()}
}
