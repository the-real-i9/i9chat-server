package helpers

import (
	"encoding/json"
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

func WSErrResp(err error, onEvent string) map[string]any {

	errCode := fiber.StatusInternalServerError

	if ferr, ok := err.(*fiber.Error); ok {
		errCode = ferr.Code
	}

	errResp := map[string]any{
		"event":   "server error",
		"onEvent": onEvent,
		"data": map[string]any{
			"statusCode": errCode,
			"errorMsg":   err.Error(),
		},
	}

	return errResp
}
