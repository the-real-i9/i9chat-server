package helpers

import (
	"encoding/json"
	"log"
	"os"

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

func Cookie(name, value, path string, maxAge int) *fiber.Cookie {
	c := &fiber.Cookie{
		HTTPOnly: true,
		Secure:   false,
		Domain:   os.Getenv("SERVER_HOST"),
	}

	c.Name = name
	c.Value = value
	c.Path = path
	c.MaxAge = maxAge

	return c
}

func AllAinB[T comparable](sA []T, sB []T) bool {
	if len(sB) == 0 {
		return false
	}

	if len(sA) == 0 {
		return true
	}

	trk := make(map[T]bool, len(sB))

	for _, el := range sB {
		trk[el] = true
	}

	for _, el := range sA {
		if !trk[el] {
			return false
		}
	}

	return true
}
