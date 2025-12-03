package helpers

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/goccy/go-json"

	"github.com/gofiber/fiber/v2"
)

func ToStruct(val any, dest any) {
	valType := reflect.TypeOf(val)

	if valType.Kind() != reflect.Map && !(valType.Kind() == reflect.Slice && valType.Elem().Kind() == reflect.Map) {
		panic("expected 'val' to be a map or slice of maps")
	}

	destType := reflect.TypeOf(dest).Elem()

	if destType.Kind() != reflect.Struct && !(destType.Kind() == reflect.Slice && destType.Elem().Kind() == reflect.Struct) {
		panic("expected 'dest' to be a pointer to struct or slice of structs")
	}

	bt, err := json.Marshal(val)
	if err != nil {
		LogError(err)
	}

	if err := json.Unmarshal(bt, dest); err != nil {
		LogError(err)
	}
}

func JoinWithCommaAnd(items ...string) string {
	n := len(items)
	if n == 0 {
		return ""
	}
	if n == 1 {
		return items[0]
	}
	if n == 2 {
		return items[0] + " and " + items[1]
	}

	// Join all except the last with commas, then append "and last"
	return strings.Join(items[:n-1], ", ") + ", and " + items[n-1]
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

func LogError(err error) {
	if err == nil {
		return
	}

	pc, file, line, ok := runtime.Caller(1)
	fn := "unknown"
	if !ok {
		file = "???"
		line = 0
	} else {
		fn = runtime.FuncForPC(pc).Name()
	}

	log.Printf("[ERROR] %s:%d %s(): %v\n", file, line, fn, err)
}

func ToJson(data any) string {
	d, err := json.Marshal(data)
	if err != nil {
		LogError(err)
	}
	return string(d)
}

func FromJson[T any](jsonStr string) (res T) {
	err := json.Unmarshal([]byte(jsonStr), &res)
	if err != nil {
		LogError(err)
	}

	return
}
