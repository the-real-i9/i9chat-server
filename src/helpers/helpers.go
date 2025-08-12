package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/gofiber/fiber/v2"
)

func ToStruct(val any, dest any) {
	destElem := reflect.TypeOf(dest).Elem()
	if destElem.Kind() != reflect.Struct && !(destElem.Kind() == reflect.Slice && destElem.Elem().Kind() == reflect.Struct) {
		panic("expected 'dest' to be a pointer to struct or slice of structs")
	}

	bt, err := json.Marshal(val)
	if err != nil {
		log.Println("helpers.go: ToStruct: json.Marshal:", err)
	}

	if err := json.Unmarshal(bt, dest); err != nil {
		log.Println("helpers.go: ToStruct: json.Unmarshal:", err)
	}
}

func WSErrReply(err error, toAction string) map[string]any {

	errCode := fiber.StatusInternalServerError

	if ferr, ok := err.(*fiber.Error); ok {
		errCode = ferr.Code
	}

	errResp := map[string]any{
		"event":    "server error",
		"toAction": toAction,
		"data": map[string]any{
			"statusCode": errCode,
			"errorMsg":   fmt.Sprint(err),
		},
	}

	return errResp
}

func WSReply(data any, toAction string) map[string]any {

	reply := map[string]any{
		"event":    "server reply",
		"toAction": toAction,
		"data":     data,
	}

	return reply
}

func Cookie(name, value string, maxAge int) *fiber.Cookie {
	c := &fiber.Cookie{
		HTTPOnly: true,
		Secure:   false,
		Path:     "/",
		Domain:   os.Getenv("SERVER_HOST"),
	}

	c.Name = name
	c.Value = value
	c.MaxAge = maxAge

	return c
}

func AsubsetB[T comparable](sA []T, sB []T) bool {
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
